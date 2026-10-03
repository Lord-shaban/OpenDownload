# Roadmap

## First release — v0.1.0

The 0.1 phase completes M0–M4, the locale/RTL portion of M5, M6 workspace
refinement and M7 public real-download deployment. Release delivery is recorded
in [issue #36](https://github.com/Lord-shaban/OpenDownload/issues/36), the
[release notes](releases/0.1.0.md) and [verification](VERIFICATION.md).
YouTube is explicitly excluded; the API and interface enforce that boundary.

### Deferred beyond 0.1

- YouTube cloud extraction: [#27](https://github.com/Lord-shaban/OpenDownload/issues/27).
  Investigation retained; experimental PR #30 is not merged.
- Dedicated public social photo-gallery adapter:
  [#13](https://github.com/Lord-shaban/OpenDownload/issues/13).
  Evaluation completed; integration and authorized live fixtures remain future work.
- Distributed workers, object storage, bulk/playlist processing and preferences
  remain product ideas, not missing 0.1 delivery gates.

## Implementation milestones

### Source expansion, current development baseline

LinkedIn public videos, Pinterest video/single-image pins and Threads public
post media are implemented in current source. Supported URL shapes, the Threads
search-preview representation and verification limits are in
[SOCIAL_SOURCES.md](SOCIAL_SOURCES.md). This does not change the published v0.1.0
release scope or complete the deferred Instagram/TikTok gallery investigation.

The M0–M4 implementation baseline passed its documented development gates.
[GitHub milestones](https://github.com/Lord-shaban/OpenDownload/milestones) and
the [issue index](GITHUB_ISSUES.md) track accepted work and remaining delivery
changes; [VERIFICATION.md](VERIFICATION.md) records evidence and release limits.
M5 retains deferred gallery integration; its locale work is accepted.
M6 and M7 are accepted. A closed release scope does not mark future features solved.

## M0 — Foundation

Research, product boundary, architecture, ADRs, design system, repository hygiene,
issue templates, labels, backlog, CI, contributor and security guidance.
Gate: review all decisions and record the real GitHub publication state.

## M1 — Safe analysis

URL/egress policy, owner-scoped metadata, bounded yt-dlp analysis, safe format
options, image detection, capability errors. Gate: adversarial URL/DNS tests,
normalization fixtures, timeouts and no arbitrary extractor flags.

## M2 — Durable downloads

SQLite jobs, worker scheduling, cancellation, retry, expiry, attachment/range
streaming, byte watchdog. Gate: queue/restart/cancel/ownership/file safety tests.

## M3 — Workspace experience

Design tokens, responsive app, paste/detect/analyze/select/queue/download states,
keyboard access, empty/error/loading states, fixture walkthrough.
Gate: typecheck/lint/build and browser flow on desktop/mobile.

## M4 — Self-hosted pre-release

Docker isolation, build/publish CI, operations runbook, privacy/security review,
real authorized extraction smoke tests. Gate: Linux container isolation and
process-group tests; dependency audit; explicit unsupported extractor evidence.

## M5 — Broader media coverage

- English/Arabic catalogs, RTL layout, mixed-direction safety and browser gates:
  implemented. See [localization](LOCALIZATION.md).
- Evaluate a dedicated public image/gallery adapter and authorized platform
  evidence: [source review and contract tests](GALLERY_EVALUATION.md) completed.
  Integration is deferred; #13 remains open for authorized live fixtures and
  challenge-stop evidence. Existing gallery ZIP processing remains bounded.

## M6 — Workspace refinement

Replace the previous rail layout with a centered, restrained glass workspace,
calm neutral/violet theme tokens and the `OpenDownload.` wordmark. Deliberate paste
starts analysis; real choices appear as needed. Include a compact supported-sites
disclosure, automatic download navigation, polished motion and useful keyboard focus.
Persist light/dark preferences and retain Arabic/RTL, actual format choices and
explicit fixture labels. Gate: production browser tests, desktop/mobile visual
review and all required GitHub checks. See issue [#20](https://github.com/Lord-shaban/OpenDownload/issues/20).

Long-term ideas and cost evaluations live in PRODUCT.md. Milestones describe work,
not delivery-date promises. No release tag before its gates pass.
