# GitHub delivery workflow

GitHub issues describe the work; branches and pull requests carry changes;
Actions provide verification evidence. The default branch is the accepted
development baseline. Releases are explicit, separately verified artifacts.

## Work lifecycle

1. Select a scoped issue, define acceptance criteria and attach its milestone
   and area/priority labels. Keep product decisions and security tradeoffs in
   repository documents or ADRs.
2. Branch from current `main`: `feat/…`, `fix/…`, `docs/…` or `chore/…`.
   Make one reviewable change. Open a draft PR early for work still in progress.
3. Link the issue with `Closes #number` when the PR satisfies its criteria.
   Include the reason for the change, meaningful test results, limitations and
   screenshots when UI behavior changes.
4. Resolve CI failures and review conversations on that branch. Refresh against
   `main` if it changes, then rerun the required checks. Mark the PR ready when
   its checks and acceptance criteria are satisfied.
5. Review the actual diff, test evidence, cancellation/resource bounds, access
   boundaries and operational implications. Squash merge the accepted PR and
   delete its branch. Update milestones when all their gates and issues are done.

Do not push implementation directly to `main`, bypass a required gate, label
authored tests as passed, or count self-review as independent approval.

## Enforced repository policy

[repository-policy.json](../.github/repository-policy.json) records the live settings:

- `main` requires a PR and an up-to-date branch.
- `go`, `web`, and `integration` must pass. Check results must come from GitHub
  Actions (app ID 15368), not an arbitrary status publisher.
- Administrators follow the same rules. Force pushes and deletion are disabled.
- Review conversations must be resolved. Stale approvals are dismissed.
- Squash merges produce a linear history; merged branches are deleted.

The repository currently has one maintainer. A second person's approval is not
an enforced gate (`required_approving_review_count: 0`); a PR, passing CI and
maintainer review are required. A solo delivery is identified as such. When a
second active maintainer joins, raise the approval requirement to one and decide
whether sensitive paths need code owners. No fictional reviewer is assigned.

With an authenticated GitHub CLI, check settings without changing them:

```sh
python scripts/github-policy.py
```

A repository administrator can reconcile the checked-in settings with
`python scripts/github-policy.py --apply`. Fork maintainers can pass
`--repo owner/repository`. Inspect the policy before applying it to another repo.
The helper checks the resulting API response; it does not inspect tokens.

## Bootstrap history

The initial implementation through `76100e1` was pushed directly to `main`.
Those commits have successful CI evidence, but were not delivered through PRs
and did not receive independent review. That published history is preserved.
Completed bootstrap issues link the actual commits and CI; retrospective PRs
are not fabricated. OD-015 establishes this enforced workflow for subsequent work.

## Releases and security

Release tags use SemVer. Before tagging, require successful CI on the exact
commit and review [VERIFICATION.md](VERIFICATION.md), capability gaps and operations
guidance. The image publication workflow checks for successful CI before pushing
versioned API/web/egress images with provenance and SBOM. The first publication
still needs a registry pull/run smoke test before a release is called verified.

Report vulnerabilities with GitHub's enabled private reporting flow. Dependabot
proposes dependency changes through PRs; those changes follow the same gates.
