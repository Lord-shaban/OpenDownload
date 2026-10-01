"""Check release identity before CI or image publication, without third-party I/O."""
import json
import os
import pathlib
import re

root = pathlib.Path(__file__).resolve().parent.parent
version = (root / "VERSION").read_text().strip()
assert re.fullmatch(r"\d+\.\d+\.\d+", version), "VERSION must use stable SemVer"
for file in ("package.json", "apps/web/package.json"):
    assert json.loads((root / file).read_text())["version"] == version, file
assert f'"version": "{version}"' in (root / "services/api/internal/server/server.go").read_text()
assert (root / "docs/releases" / f"{version}.md").is_file(), "Release notes missing"
if os.environ.get("GITHUB_REF_TYPE") == "tag":
    assert os.environ["GITHUB_REF_NAME"] == "v" + version, "Tag and VERSION disagree"
print(f"Release metadata agrees: v{version}")
