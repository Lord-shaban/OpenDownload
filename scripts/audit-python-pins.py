"""Check the pinned browser provider dependencies against public PyPI advisories."""
import json
import pathlib
import re
import sys
import urllib.request

for line in pathlib.Path(sys.argv[1]).read_text(encoding="utf8").splitlines():
    line = line.strip()
    if not line or line.startswith("#"):
        continue
    if not re.fullmatch(r"[A-Za-z0-9_.-]+==[A-Za-z0-9.]+", line):
        raise SystemExit("Every browser dependency must be pinned.")
    name, version = line.split("==")
    with urllib.request.urlopen(f"https://pypi.org/pypi/{name}/{version}/json", timeout=30) as response:
        metadata = json.load(response)
    vulnerabilities = metadata.get("vulnerabilities", [])
    if vulnerabilities:
        raise SystemExit(f"{name} {version} has reported vulnerabilities; update before building.")
    print(f"{name} {version}: no reported PyPI vulnerabilities")
