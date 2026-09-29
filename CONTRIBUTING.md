# Contributing

## Working agreement

1. Pick a scoped issue from the backlog. Explain the user problem before proposing
   another dependency. Security changes need a threat-model update.
2. Create a branch (`feat/…`, `fix/…`, `docs/…`, `chore/…`). Keep one logical change per PR.
3. Add meaningful tests for changed behavior. Use deterministic fake extractors
   in CI; do not rely on real platforms for required tests.
4. Run the commands in `docs/TESTING.md`, format Go with gofmt and frontend with
   Prettier, and update documentation alongside behavior changes.
5. Use a descriptive conventional commit (`feat:`, `fix:`, `docs:`, `test:`,
   `chore:`). Link the issue and include verification evidence in the PR.

Never submit credentials, real private URLs, downloaded copyrighted media, or
user logs. Screenshots must use fixtures or media you have permission to share.
Do not add cookie import, authentication bypass, watermark removal, arbitrary
shell commands, arbitrary paths, or unbounded batch downloading.

## Review

`main` is protected, including for administrators. PRs must pass the `go`, `web`
and `integration` checks, be up to date and resolve review conversations before
squash merge. See [GitHub workflow](docs/GITHUB_WORKFLOW.md) for the enforced settings,
single-maintainer review policy and release requirements.

Reviewers check correctness, resource bounds, cancellation, access boundaries,
accessibility, and behavior on mobile. Dependencies must explain their cost and
purpose. Breaking API changes require explicit migration notes.

Until releases exist, main is the development branch. Release tags follow SemVer.
Do not call an unverified pre-release production-ready. Governance is initially
maintainer-led; substantive decisions are captured as ADRs.
