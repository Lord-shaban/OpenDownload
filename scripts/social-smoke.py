"""Opt-in public-source download evidence; third-party uptime is not a CI gate."""
import argparse
import hashlib
import http.cookiejar
import json
import pathlib
import subprocess
import time
import urllib.error
import urllib.request
import zipfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--base-url", default="http://127.0.0.1:8080")
parser.add_argument("--origin", default="http://localhost:3000")
parser.add_argument("--source", required=True)
parser.add_argument("--kind", choices=("video", "image", "audio"), default="video")
parser.add_argument("--ffprobe", default="ffprobe")
parser.add_argument("--output", default=".data/social-smoke-output")
args = parser.parse_args()
opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))


def api(path, body=None, method=None):
    request = urllib.request.Request(
        args.base_url + "/api/v1" + path, method=method,
        data=None if body is None else json.dumps(body).encode(),
        headers={"Origin": args.origin, "Content-Type": "application/json"},
    )
    with opener.open(request, timeout=60) as response:
        data = response.read()
        return json.loads(data) if data else None


assert not api("/status")["fixtureMode"], "Use a real engine, not fixtures."
analysis = api("/analyze", {"url": args.source})
option = next((o for o in analysis["options"] if o["kind"] == args.kind), None)
assert option, f"No {args.kind} returned for this public sample."
job = api("/jobs", {"analysisId": analysis["id"], "optionId": option["id"]})
try:
    deadline = time.monotonic() + 300
    while job["state"] in ("queued", "processing"):
        assert time.monotonic() < deadline, "Download timed out."
        time.sleep(0.5)
        job = api("/jobs/" + job["id"])
    assert job["state"] == "complete", job.get("error", job["state"])
    endpoint = args.base_url + "/api/v1/jobs/" + job["id"] + "/files/0"
    with opener.open(urllib.request.Request(endpoint, headers={"Range": "bytes=0-9"}), timeout=15) as response:
        assert response.status == 206 and len(response.read()) == 10, "Range serving failed."
    output = pathlib.Path(args.output)
    output.mkdir(parents=True, exist_ok=True)
    destination = output / (analysis["platform"].lower() + "." + option["extension"])
    size = 0
    digest = hashlib.sha256()
    with opener.open(endpoint, timeout=60) as response, destination.open("wb") as file:
        while chunk := response.read(65536):
            size += len(chunk)
            assert size <= 128 << 20, "Smoke sample exceeds 128 MiB."
            digest.update(chunk)
            file.write(chunk)
    assert size == job["files"][0]["bytes"] and size > 0, "Attachment size mismatch."
    result = {"platform": analysis["platform"], "selection": option["id"], "bytes": size, "sha256": digest.hexdigest()}
    if option["extension"] == "zip":
        with zipfile.ZipFile(destination) as archive:
            assert archive.testzip() is None and 1 <= len(archive.infolist()) <= 20
            for entry in archive.infolist():
                assert not entry.is_dir() and pathlib.PurePosixPath(entry.filename).name == entry.filename
                with archive.open(entry) as image:
                    head = image.read(16)
                assert head.startswith((b"\xff\xd8\xff", b"\x89PNG\r\n\x1a\n", b"GIF8")) or head[:4] == b"RIFF" and head[8:12] == b"WEBP", "Gallery contains non-image data."
            result["images"] = len(archive.infolist())
    else:
        probe = json.loads(subprocess.check_output([
            args.ffprobe, "-v", "error", "-show_streams", "-show_format", "-of", "json", str(destination),
        ]))
        kinds = [stream["codec_type"] for stream in probe["streams"]]
        assert ("video" if args.kind == "image" else args.kind) in kinds, "Unexpected media type."
        result["streams"] = kinds
        result["duration"] = probe["format"].get("duration")
    print(json.dumps(result))
finally:
    api("/jobs/" + job["id"], method="DELETE")
try:
    with opener.open(endpoint, timeout=15):
        raise AssertionError("Deleted file is still accessible.")
except urllib.error.HTTPError as error:
    assert error.code == 404, error.code
print("Real download, attachment bytes, range serving and deletion passed.")
