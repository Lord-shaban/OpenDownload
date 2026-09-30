# Verification record

Date: 2026-09-29. Verified implementation baseline: [`76100e1`](https://github.com/Lord-shaban/OpenDownload/commit/76100e1).
The complete [Linux CI run](https://github.com/Lord-shaban/OpenDownload/actions/runs/36584693905)
passed all three jobs: `go`, `web`, and `integration`.

| Stage                  | State              | Evidence                                                                                                   |
| ---------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------- |
| M0 research/design     | Recorded           | Dated primary sources, ADRs, UI UX Pro Max query and product design tokens                                 |
| GitHub foundation      | Published          | Public repository, MIT license, contributor/security policies, templates, six milestones and issue index   |
| M1 analysis/security   | Passed             | Go policy, proxy, normalization, ownership, bounds and HTTP tests in Linux CI                              |
| M2 downloads/storage   | Passed             | Queue/restart/race, cancellation, expiry, scratch budget, output validation and range tests                |
| M3 workspace           | Passed             | Lint, TypeScript, four unit tests, production build and ten Chromium E2E tests                             |
| M4 container isolation | Passed in Linux CI | All three images built; Compose healthy; API direct TCP egress blocked; proxy refused private destinations |
| Real media smoke       | Passed on Windows  | Public Blender trailer: actual MP4, FFmpeg MP3 conversion, ffprobe and HTTP 206 ranges                     |

## Deterministic gates

The workflow in [ci.yml](../.github/workflows/ci.yml) ran:

- `gofmt`, `go vet ./...`, `go test -race -count=1 ./...` and `govulncheck ./...`.
  Linux tests include termination of subprocess descendants, cancellation cleanup,
  DNS pinning, unsafe redirects and owner access failures.
- `pnpm --filter web lint`, `typecheck`, `test`, `build`, and
  `pnpm audit --prod --audit-level high`. Both dependency audits passed on this date.
- Five browser scenarios on desktop and mobile: analysis/selection/queue/save/delete,
  cancellation/retry, invalid/unsupported URLs, processing failure/retry, and
  reflow/reduced motion/dialog keyboard focus. Ten tests passed. The saved fixture
  file's contents are checked; fixtures produce labeled text, not pretend media.
- `docker compose build`, fixture-mode `docker compose up -d --wait` and
  `python3 scripts/container-check.py`. The API could not open a direct connection
  to `1.1.1.1:443`; the guarded proxy refused metadata/link-local and loopback targets.
  Docker was unavailable locally; these checks ran on GitHub's Linux runner.

The web container has a separate bridge for its published loopback port. The
network fence applies to the API and extractors. See [ADR 0002](adr/0002-guarded-egress.md).

## Real media evidence

[media-smoke.py](../scripts/media-smoke.py) ran against a native Go API with
`OD_FIXTURE_MODE=false`, using the guarded proxy, yt-dlp 2026.08.19 and FFmpeg 9.0.2.
Source: [public Sintel trailer](https://download.blender.org/durian/trailer/sintel_trailer-480p.mp4).

| Output        | Bytes     | ffprobe streams | Duration    |
| ------------- | --------- | --------------- | ----------- |
| Original MP4  | 4,372,373 | Video and audio | 52.208333 s |
| Converted MP3 | 1,247,953 | Audio           | 51.925333 s |

Both attachments returned HTTP 206 for `Range: bytes=0-9` with exactly ten bytes.
The browser also completed a real MP4 job through the same-origin web API.
Downloaded samples and job data are ignored and are not committed.

To repeat with the real tools configured:

```sh
python scripts/media-smoke.py --base-url http://127.0.0.1:8080 --origin http://localhost:3000
```

Manual preview checks covered light/dark themes, mobile layout, visible errors and
keyboard interactions. The [workspace screenshot](assets/workspace.png) shows the
development preview and a completed public-trailer job. This is not a WCAG certification.

## GitHub delivery controls

The repository now requires PRs, up-to-date branches, successful `go`, `web` and
`integration` checks from GitHub Actions, and resolved review conversations.
Rules apply to administrators; force pushes and deletion of `main` are disabled.
Squash is the permitted merge method. Private vulnerability reporting is enabled.
The live configuration was read back with `python scripts/github-policy.py`.
See [the delivery workflow](GITHUB_WORKFLOW.md) for the single-maintainer review policy.

## Remaining coverage and release limits

- CI fixtures do not establish live YouTube, Instagram, TikTok, X, Facebook,
  Reddit, Vimeo or SoundCloud availability. Real smoke coverage is a public
  direct MP4 and audio conversion; live subtitles, thumbnails and galleries
  have not been checked across that platform matrix.
- Dedicated public photo/gallery extraction remains under M5 evaluation.
  Existing bounded extractor image entries are covered by normalization tests.
- Image publication is authored and gated, but no version tag, GitHub release,
  GHCR publication or published-image pull test has run. This record qualifies
  the development baseline, not a published release.
- Restore operations, sustained load, host disk quota enforcement and hostile
  multi-tenant deployment need operator validation. Containers do not replace
  stronger worker isolation for an anonymous public service.
- Dependency locks and pinned tool versions help repeat builds; mutable base-image
  tags and distribution packages do not guarantee byte-identical images.
- No independent human security review has occurred.

## Arabic/RTL verification — 2026-09-29

[PR #17](https://github.com/Lord-shaban/OpenDownload/pull/17) merged after
[CI run 36592605343](https://github.com/Lord-shaban/OpenDownload/actions/runs/36592605343)
passed all required jobs on head `9f93d125c8c36bddc87f69b42fe5aa83525d8f38`.

Local checks passed: ESLint, TypeScript, eight Vitest tests, Go URL security tests
and sixteen Chromium browser tests (eight scenarios on desktop and mobile).
Browser tests now build and run the production standalone application, including
its copied static/public assets, instead of the development server.

The additional scenarios verify language persistence and server-rendered RTL,
preserving an entered URL while switching, Arabic keyboard format selection,
queue/save/delete with fixture file content checks, translated URL errors,
mixed-direction rejection, dialog keyboard dismissal/focus, reduced motion,
375–1440px reflow and same-origin font requests. These fixtures test application
behavior; they do not establish live platform availability.

Manual review also covered Arabic light/dark desktop/mobile and a translated 404.
The [Arabic preview](assets/workspace-ar.png) shows the development workspace.

## Gallery contract verification — 2026-09-29

Six additional Go test functions passed locally, including table cases for both
synthetic platform labels, restricted/mixed collections and download failures.
Original PNGs are encoded during tests; ZIP members preserve exact bytes and
decode successfully. HTTP archive requests use a local proxy; the actual guarded
proxy rejects loopback/metadata assets. Tests also verify the 20-item bound,
aggregate byte refusal, 403 termination and cancellation before the next image.

No gallery-dl package or runtime path was added. [Source review and coverage
matrix](GALLERY_EVALUATION.md) explain the decision. No live Instagram/TikTok photo
extraction was run; no captured platform fixture is claimed.

## Minimal workspace revision — 2026-09-29

Issue [#20](https://github.com/Lord-shaban/OpenDownload/issues/20) tracks the visual
redesign and the user's refinement toward a minimal cobalt-like interaction.
The header/marketing copy and visible stepper were removed from the primary
flow. Deliberate clipboard/native paste now analyzes immediately; typing still
waits for Enter/the arrow. Real media choices and explicit fixture identity
remain intact. English/Arabic API errors now use the same safe catalog rather
than raw diagnostics.

New browser scenarios cover clipboard/native paste without an extra submit,
clipboard denial and keyboard recovery, and persistent/server-rendered theme
with focus through Enter → formats → Downloads → new link. Arabic reflow includes
320px. Visual artifacts are actual previews, including
[Arabic dark](assets/workspace-glass-ar-dark.jpg),
[English light](assets/workspace-glass-en-light.jpg),
[Arabic mobile](assets/workspace-glass-ar-mobile.jpg) and
[supported sites on mobile](assets/workspace-glass-sites-mobile.jpg).

## Owned direct image smoke — 2026-09-29

A real API with fixture mode disabled downloaded this repository's original
[public preview asset](https://raw.githubusercontent.com/Lord-shaban/OpenDownload/588a4adbd401e5931c5f64f38a599f7acd3f4ff5/docs/assets/workspace-ar.png)
through the guarded proxy. The old asset has a `.png` filename but JPEG bytes;
the downloader correctly produced `media.jpg` with the exact 47,960 bytes.
SHA-256: `6c09b99393308a77601defe19ec031d28d0f30209e293282124b200a59a0e6e8`.
Range `bytes=0-9` returned HTTP 206 and ten bytes; deleting the job revoked the
attachment with HTTP 404. New preview artifacts use the correct `.jpg` suffix.
This verifies the direct-image pipeline, not live Instagram/TikTok photo access.

Final local gates for this revision passed: ESLint, strict TypeScript, eight unit
checks and twenty-two Chromium tests (eleven scenarios on desktop/mobile) against
the production standalone build. The icon dependency is pinned and imports only
used modules. See the redesign PR for checks on the exact committed head.
