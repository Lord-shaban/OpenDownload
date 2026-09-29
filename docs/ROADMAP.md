# Roadmap

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

Dedicated public image/gallery adapter; platform-specific authorized fixtures,
gallery zip with item limit, locale support. Evaluate user value before extras.

Long-term ideas and cost evaluations live in PRODUCT.md. Milestones describe work,
not delivery-date promises. No release tag before its gates pass.
