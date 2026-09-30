"""Verify the actual preview image, including restart persistence and shutdown."""
import http.cookiejar
import json
import subprocess
import time
import urllib.error
import urllib.request

base = "http://127.0.0.1:3002"
name = "opendownload-preview-check"
volume = name + "-data"
client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))


def docker(*args):
    return subprocess.check_output(["docker", *args], text=True).strip()


def call(path, body=None):
    request = urllib.request.Request(base + path, data=None if body is None else json.dumps(body).encode())
    request.add_header("Origin", base)
    request.add_header("Content-Type", "application/json")
    return client.open(request, timeout=5)


def ready():
    deadline = time.monotonic() + 45
    while time.monotonic() < deadline:
        try:
            with call("/api/v1/status") as response:
                status = json.load(response)
            assert status["ready"] and status["fixtureMode"], status
            assert not any(status["dependencies"].values()), status
            return
        except (urllib.error.URLError, OSError):
            time.sleep(0.5)
    raise AssertionError("Preview did not become ready")


try:
    docker("run", "--detach", "--name", name, "--publish", "127.0.0.1:3002:3000",
           "--env", "OD_ORIGIN=" + base, "--env", "OD_FIXTURE_MODE=false",
           "--mount", "type=volume,src=" + volume + ",dst=/data", "opendownload-preview:check")
    ready()
    with call("/") as response:
        assert b"OpenDownload" in response.read()
    with call("/api/v1/analyze", {"url": "https://example.com/sample"}) as response:
        analysis = json.load(response)
    with call("/api/v1/jobs", {"analysisId": analysis["id"], "optionId": "video-1080"}) as response:
        assert response.status == 202
        job = json.load(response)
    deadline = time.monotonic() + 20
    while time.monotonic() < deadline:
        with call("/api/v1/jobs/" + job["id"]) as response:
            job = json.load(response)
        if job["state"] == "complete":
            break
        time.sleep(0.3)
    assert job["state"] == "complete", job
    file_path = "/api/v1/jobs/" + job["id"] + "/files/0"
    with call(file_path) as response:
        content = response.read()
        assert b"not extracted media" in content
        assert "attachment" in response.headers["Content-Disposition"]
    docker("stop", "--time", "20", name)
    assert docker("inspect", "--format", "{{.State.ExitCode}}", name) == "0"
    docker("start", name)
    ready()
    with call(file_path) as response:
        assert response.read() == content, "Fixture data did not survive restart"
    ports = json.loads(docker("inspect", "--format", "{{json .HostConfig.PortBindings}}", name))
    assert set(ports) == {"3000/tcp"}, ports
    print("Preview image: forced fixture mode, UI/API flow, file persistence and graceful shutdown passed.")
finally:
    subprocess.run(["docker", "logs", "--tail=40", name], check=False)
    subprocess.run(["docker", "rm", "--force", name], check=False)
    subprocess.run(["docker", "volume", "rm", volume], check=False)
