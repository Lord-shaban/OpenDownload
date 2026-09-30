# Roadmap

The M0–M4 implementation baseline has passed its documented development gates.
[GitHub milestones](https://github.com/Lord-shaban/OpenDownload/milestones) and
the [issue index](GITHUB_ISSUES.md) track accepted work and remaining delivery
changes; [VERIFICATION.md](VERIFICATION.md) records evidence and release limits.
No versioned release has been published. M5 tracks gallery integration;
M6 tracks the requested workspace redesign.

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
