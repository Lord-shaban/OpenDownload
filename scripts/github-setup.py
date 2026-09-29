"""Create missing GitHub planning metadata using an authenticated gh CLI.

Usage: python scripts/github-setup.py [--repo owner/name]
Never changes existing issues' state. No tokens are read or logged.
"""
import argparse
import json
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]


def api(path, method="GET", payload=None):
    args = ["gh", "api", path, "--method", method]
    if payload is not None:
        args += ["--input", "-"]
    result = subprocess.run(args, input=json.dumps(payload) if payload is not None else None,
                            text=True, capture_output=True, check=True)
    return json.loads(result.stdout) if result.stdout.strip() else None


def main():
    plan = json.loads((ROOT / "scripts/github-plan.json").read_text())
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", default=plan["repository"])
    opts = parser.parse_args()
    base = f"repos/{opts.repo}"
    labels = {item["name"] for item in api(f"{base}/labels?per_page=100")}
    for name, color in plan["labels"].items():
        if name not in labels:
            api(f"{base}/labels", "POST", {"name": name, "color": color})
    milestones = {item["title"]: item["number"] for item in api(f"{base}/milestones?state=all&per_page=100")}
    for item in plan["milestones"]:
        if item["title"] not in milestones:
            created = api(f"{base}/milestones", "POST", item)
            milestones[item["title"]] = created["number"]
    issues = {}
    page = 1
    while True:
        batch = api(f"{base}/issues?state=all&per_page=100&page={page}")
        issues.update({item["title"]: item for item in batch if "pull_request" not in item})
        if len(batch) < 100:
            break
        page += 1
    links = []
    for issue in plan["issues"]:
        title = f"{issue['key']}: {issue['title']}"
        if title not in issues:
            body = "## Acceptance criteria\n\n" + "\n".join(f"- [ ] {a}" for a in issue["acceptance"])
            body += "\n\n## References\n\nSee docs/BACKLOG.md, docs/TESTING.md and related ADRs."
            issues[title] = api(f"{base}/issues", "POST", {
                "title": title, "body": body,
                "labels": [f"area:{issue['area']}", "priority:high" if issue["milestone"] < 3 else "priority:normal"],
                "milestone": milestones[plan["milestones"][issue["milestone"]]["title"]],
            })
        links.append(f"| {issue['key']} | [#{issues[title]['number']}]({issues[title]['html_url']}) | {issue['title']} |")
    (ROOT / "docs/GITHUB_ISSUES.md").write_text(
        "# GitHub issue index\n\n| Key | Issue | Scope |\n|---|---|---|\n" + "\n".join(links) + "\n", encoding="utf-8")
    print(f"GitHub plan synchronized: {opts.repo}; {len(links)} issues.")


if __name__ == "__main__":
    main()
