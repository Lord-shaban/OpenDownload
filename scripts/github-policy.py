"""Check GitHub delivery policy; use --apply to reconcile the checked-in settings.

Requires an authenticated gh CLI. Reads no tokens and never rewrites Git history.
Run from any directory: python scripts/github-policy.py [--repo owner/name] [--apply]
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
    result = subprocess.run(
        args, input=None if payload is None else json.dumps(payload),
        text=True, encoding="utf-8", capture_output=True, check=True,
    )
    return json.loads(result.stdout)


def subset_matches(actual, expected):
    if isinstance(expected, dict):
        return isinstance(actual, dict) and all(
            (key in actual or value is None) and subset_matches(actual.get(key), value)
            for key, value in expected.items()
        )
    return actual == expected


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", default="Lord-shaban/OpenDownload")
    parser.add_argument("--apply", action="store_true")
    opts = parser.parse_args()
    policy = json.loads((ROOT / ".github/repository-policy.json").read_text(encoding="utf-8"))
    base = f"repos/{opts.repo}"
    if opts.apply:
        api(base, "PATCH", policy["repository"])
        api(f"{base}/branches/main/protection", "PUT", policy["main"])
    repository = api(base)
    protection = api(f"{base}/branches/main/protection")
    expected = dict(policy["main"])
    expected["required_status_checks"] = dict(expected["required_status_checks"])
    # GitHub includes the check names in contexts even when checks were supplied.
    expected["required_status_checks"].pop("contexts", None)
    expected_checks = sorted(expected["required_status_checks"].pop("checks"), key=lambda c: c["context"])
    actual_checks = sorted(protection["required_status_checks"]["checks"], key=lambda c: c["context"])
    for key, value in list(expected.items()):
        if isinstance(value, bool):
            expected[key] = {"enabled": value}
    if (not subset_matches(repository, policy["repository"])
            or not subset_matches(protection, expected)
            or actual_checks != expected_checks):
        raise SystemExit("GitHub delivery policy differs from .github/repository-policy.json.")
    print(f"Verified {opts.repo}: protected main, required GitHub Actions checks, PRs and squash merges.")


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise SystemExit(error.stderr.strip() or "GitHub API request failed.") from None
