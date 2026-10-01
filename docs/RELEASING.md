# Release procedure

The first supported line is 0.1.x. A release is a fixed verified scope; future
features remain explicitly deferred instead of being called complete.

1. Create a release issue/milestone and a scoped PR from current `main`.
2. Set `VERSION`, root/web package versions and API status version together.
   Add `CHANGELOG.md` and `docs/releases/<version>.md`; update capabilities,
   support policy and roadmap. `scripts/release-check.py` checks identity.
3. Run meaningful local checks and all required GitHub CI. Review the actual
   diff and evidence; identify solo review. Merge normally, without bypassing
   protected-branch gates.
4. Wait for successful CI on the merged main commit. Create an annotated
   `v<version>` tag on that exact commit and push it. Never move a published tag.
5. The publication workflow verifies tag/version/CI, publishes API/web/egress/cloud
   images with source/revision/license labels, provenance and SBOM, then pulls
   and tests those images. Check its actual result before publishing the release.
6. Make the new GHCR packages public through their package settings, then verify
   anonymous access. GitHub initially creates container packages as private;
   repository access inheritance does not imply public package visibility.
   [GitHub package visibility](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility).
7. Deploy the accepted commit, verify version/real-mode readiness, release boundary
   messages and actual permitted media. Retain `/data`; remove temporary diagnostics.
8. Publish GitHub release notes and record exact CI/publication/deployment evidence
   in the release issue. Close its milestone only after these gates complete.

Failures keep the release draft. Image upload alone is not a published-image
runtime test. Fixtures, local-only extraction, analysis-only success and experimental
branches do not prove a real cloud download. Do not fabricate approval or security
certification. See [GitHub workflow](GITHUB_WORKFLOW.md).
