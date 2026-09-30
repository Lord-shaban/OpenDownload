"""Linux smoke of real extraction inside the single-container network fence."""
import hashlib
import http.cookiejar
import json
import os
import pathlib
import subprocess
import time
import urllib.error
import urllib.request

base = "http://127.0.0.1:3003"
name = "opendownload-cloud-check"
volume = name + "-data"
client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))


def docker(*args):
    return subprocess.check_output(["docker", *args], text=True).strip()


def call(path, body=None, method=None, opener=client, headers=None):
    request = urllib.request.Request(base + path, method=method,
        data=None if body is None else json.dumps(body).encode(),
        headers={"Origin": base, "Content-Type": "application/json", **(headers or {})})
    return opener.open(request, timeout=45)


def api(path, body=None, method=None):
    with call("/api/v1" + path, body, method) as response:
        data = response.read()
        return json.loads(data) if data else None


def ready():
    deadline = time.monotonic() + 75
    while time.monotonic() < deadline:
        try:
            status = api("/status")
            assert status["ready"] and not status["fixtureMode"], status
            assert all(status["dependencies"].values()), status
            assert status["limits"]["maxBytes"] == 134217728, status
            return
        except (urllib.error.URLError, OSError):
            time.sleep(0.5)
    raise AssertionError("Cloud did not become ready")


def complete(job):
    deadline = time.monotonic() + 90
    while job["state"] in ("queued", "processing"):
        assert time.monotonic() < deadline, "Job timeout"
        time.sleep(0.3)
        job = api("/jobs/" + job["id"])
    assert job["state"] == "complete", job
    return job


def refused(path, opener=client, body=None):
    try:
        with call(path, body=body, opener=opener):
            raise AssertionError("Access unexpectedly allowed: " + path)
    except urllib.error.HTTPError as error:
        assert error.code in (400, 403, 404, 422), error.code


try:
    docker("run", "--detach", "--name", name, "--publish", "127.0.0.1:3003:3000",
           "--user", "10001:10001", "--cap-drop", "ALL", "--read-only",
           "--security-opt", "no-new-privileges:true", "--security-opt", "seccomp:unconfined",
           "--security-opt", "apparmor:unconfined", "--pids-limit", "96", "--memory", "512m",
           "--tmpfs", "/tmp:size=64m,uid=10001,gid=10001", "--env", "OD_ORIGIN=" + base,
           "--env", "OD_FIXTURE_MODE=true", "--mount", "type=volume,src=" + volume + ",dst=/data",
           "opendownload-cloud:check")
    ready()
    with call("/") as response:
        assert b"OpenDownload" in response.read()
    check = """
import http.client, socket
c=http.client.HTTPConnection('worker',timeout=5)
c.sock=socket.socket(socket.AF_UNIX)
c.sock.settimeout(5)
c.sock.connect('/tmp/opendownload-cloud/check.sock')
c.request('GET','/check')
r=c.getresponse()
assert r.status==200, r.read()
assert b'direct egress denied' in r.read()
print('Isolated worker denies direct internet and private proxy destinations.')
"""
    docker("exec", name, "python3", "-c", check)
    ports = json.loads(docker("inspect", "--format", "{{json .HostConfig.PortBindings}}", name))
    assert set(ports) == {"3000/tcp"}, ports
    assert docker("exec", name, "id", "-u") == "10001"
    for source in ("http://169.254.169.254/latest/meta-data/", "https://127.0.0.1/"):
        refused("/api/v1/analyze", body={"url": source})
    ref = os.environ.get("OD_TEST_MEDIA_REF") or subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    source = "https://raw.githubusercontent.com/Lord-shaban/OpenDownload/" + ref + "/tests/assets/sample.mp4"
    analysis = api("/analyze", {"url": source})
    video = next(o for o in analysis["options"] if o["kind"] == "video")
    audio = next(o for o in analysis["options"] if o["id"] == "audio-mp3")
    other = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    saved = []
    for option in (video, audio):
        job = complete(api("/jobs", {"analysisId": analysis["id"], "optionId": option["id"]}))
        path = "/api/v1/jobs/" + job["id"] + "/files/0"
        refused(path, opener=other)
        with call(path, headers={"Range": "bytes=0-9"}) as response:
            assert response.status == 206 and len(response.read()) == 10
        with call(path) as response:
            data = response.read(1048577)
            assert len(data) <= 1048576 and "attachment" in response.headers["Content-Disposition"]
        probe = json.loads(subprocess.check_output(["docker", "exec", "-i", name, "ffprobe", "-v", "error",
            "-show_streams", "-show_format", "-of", "json", "-i", "pipe:0"], input=data))
        kinds = [s["codec_type"] for s in probe["streams"]]
        assert option["kind"] in kinds and float(probe["format"]["duration"]) > 1.5, probe
        if option["kind"] == "video":
            assert hashlib.sha256(data).digest() == hashlib.sha256(pathlib.Path("tests/assets/sample.mp4").read_bytes()).digest()
        saved.append((job["id"], path, data))
    docker("stop", "--time", "20", name)
    assert docker("inspect", "--format", "{{.State.ExitCode}}", name) == "0"
    docker("start", name)
    ready()
    docker("exec", name, "python3", "-c", check)
    for job_id, path, expected in saved:
        with call(path) as response:
            assert response.read() == expected, "Media or owner access did not survive restart"
        api("/jobs/" + job_id, method="DELETE")
        refused(path)
    # Removing namespace tooling must fail closed before any public listener.
    failed = docker("run", "--detach", "--name", name + "-failure", "--entrypoint", "/usr/local/bin/opendownload-cloud", "--env", "PATH=/missing",
                    "opendownload-cloud:check")
    assert docker("wait", failed) == "1"
    logs = docker("logs", failed)
    assert "real cloud downloads ready" not in logs
    print("Cloud: real video/MP3, network fence, private URL refusal, ownership, ranges, persistence, fail-closed startup and shutdown passed.")
finally:
    subprocess.run(["docker", "logs", "--tail=45", name], check=False)
    subprocess.run(["docker", "rm", "--force", name, name + "-failure"], check=False)
    subprocess.run(["docker", "volume", "rm", volume], check=False)
