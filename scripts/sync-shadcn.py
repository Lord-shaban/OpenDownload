"""Copy only used components from the official registry after CLI failure.

Source imports are adapted the same way the CLI adapts registry source imports.
Run deliberately when upgrading; this never runs automatically at build time.
"""
import json
import pathlib
import urllib.request

root = pathlib.Path(__file__).resolve().parents[1] / "apps/web/src/components/ui"
root.mkdir(parents=True, exist_ok=True)
for item in ("button", "dialog"):
    with urllib.request.urlopen(f"https://ui.shadcn.com/r/styles/new-york-v4/{item}.json", timeout=30) as response:
        data = json.load(response)
    for file in data["files"]:
        content = file["content"].replace('from "cn"', 'from "@/lib/utils"')
        content = content.replace('from "@/registry/new-york-v4/ui/button"', 'from "@/components/ui/button"')
        content = content.replace('from "@/registry/new-york/ui/button"', 'from "@/components/ui/button"')
        if item == "button":
            content = content.replace('icon: "size-9"', 'icon: "size-11"')
        (root / pathlib.Path(file["path"]).name).write_text(content, encoding="utf-8")
    print(f"Copied {item} from the official shadcn registry.")
