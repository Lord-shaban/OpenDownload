"""Opt-in real public media smoke. Never used as a third-party uptime gate in CI."""
import argparse
import http.cookiejar
import json
import os
import pathlib
import subprocess
import time
import urllib.error
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument("--base-url", default="http://127.0.0.1:8080")
parser.add_argument("--origin", default="http://localhost:3000")
parser.add_argument("--source", default="https://download.blender.org/durian/trailer/sintel_trailer-480p.mp4")
parser.add_argument("--ffprobe", default="ffprobe")
parser.add_argument("--output", default=".data/smoke")
args = parser.parse_args()
handlers = [urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())]
access_user = os.environ.get("OD_SMOKE_USER")
access_password = os.environ.get("OD_SMOKE_PASSWORD")
if access_user or access_password:
    if not access_user or not access_password:
        parser.error("Set both OD_SMOKE_USER and OD_SMOKE_PASSWORD for protected hosting.")
    passwords = urllib.request.HTTPPasswordMgrWithDefaultRealm()
    passwords.add_password(None, args.base_url, access_user, access_password)
    handlers.append(urllib.request.HTTPBasicAuthHandler(passwords))
opener = urllib.request.build_opener(*handlers)

def api(path, body=None, method=None):
    request = urllib.request.Request(args.base_url + "/api/v1" + path, method=method,
        data=None if body is None else json.dumps(body).encode(),
        headers={"Origin": args.origin, "Content-Type": "application/json"})
    with opener.open(request, timeout=50) as response:
        data = response.read()
        return json.loads(data) if data else None

status = api("/status")
assert status["ready"] and not status["fixtureMode"], "Real media tools must be enabled."
analysis = api("/analyze", {"url": args.source})
output = pathlib.Path(args.output)
output.mkdir(parents=True, exist_ok=True)
for option in (next(o for o in analysis["options"] if o["kind"] == "video"),
               next(o for o in analysis["options"] if o["id"] == "audio-mp3")):
    job = api("/jobs", {"analysisId": analysis["id"], "optionId": option["id"]})
    deadline = time.monotonic() + 180
    while job["state"] in ("queued", "processing"):
        assert time.monotonic() < deadline, "Job deadline exceeded."
        time.sleep(0.5)
        job = api("/jobs/" + job["id"])
    assert job["state"] == "complete", job.get("error", job["state"])
    endpoint = args.base_url + "/api/v1/jobs/" + job["id"] + "/files/0"
    with opener.open(urllib.request.Request(endpoint, headers={"Range":"bytes=0-9"}), timeout=10) as response:
        assert response.status == 206 and len(response.read()) == 10, "Range serving failed."
    destination = output / ("public-trailer." + option["extension"])
    total = 0
    with opener.open(endpoint, timeout=60) as response, destination.open("wb") as file:
        while chunk := response.read(65536):
            total += len(chunk)
            assert total <= 20 << 20, "Smoke sample exceeded 20 MiB."
            file.write(chunk)
    probe = json.loads(subprocess.check_output([args.ffprobe, "-v", "error", "-show_streams", "-show_format", "-of", "json", str(destination)]))
    kinds = [stream["codec_type"] for stream in probe["streams"]]
    assert option["kind"] in kinds, kinds
    assert float(probe["format"]["duration"]) > 1, "Invalid media duration."
    print(json.dumps({"selection": option["id"], "bytes": total, "streams": kinds, "duration": probe["format"]["duration"]}))
    api("/jobs/" + job["id"], method="DELETE")
    try:
        with opener.open(endpoint, timeout=10):
            raise AssertionError("Deleted job attachment is still accessible.")
    except urllib.error.HTTPError as error:
        assert error.code == 404, error.code
print("Real video, audio conversion, range streaming and deletion passed.")
