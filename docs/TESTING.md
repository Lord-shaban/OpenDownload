# Coding and testing standards

Go: standard library HTTP router, context propagation, small interfaces at external
boundaries, explicit error returns, slog structured fields, gofmt/go vet. No global
mutable application state. SQLite writes stay short. Cancellation and terminal
state transitions must remain correct under races. Use parameterized SQL only.

Frontend: strict TypeScript, ESLint, Prettier, functional React, server shell/client
workspace boundary, semantic controls, source-owned shadcn primitives. Keep API
types together; treat network JSON as untrusted. No user-visible raw exceptions.
Only introduce dependencies that solve a documented problem.

## Required gates

| Area | Evidence |
|---|---|
| URL/egress | Special-use IPs, mixed DNS, rebinding-safe pinned addresses, disallowed schemes/ports, credentials |
| Metadata | DRM/live/auth policy, safe IDs, duplicate qualities, video/audio distinction, missing capabilities |
| Queue | Capacity, owner scope, atomic claims, terminal transitions, restart recovery |
| Worker/storage | Cancel, timeouts, subprocess failure, malicious output path/symlink, cleanup, byte limits |
| HTTP | JSON limits, CSRF/origin, cookies, hidden ownership failures, range/attachment serving |
| Frontend | URL detection/format utilities, lint, strict typecheck, production build |
| E2E | Fixture analysis → select → create → complete → save, failure/retry/cancel, narrow-screen reflow |
| Linux | Race detector, process descendants killed, direct egress blocked, Docker health |

Tests assert behavior and failures, not markup trivia. Use test temporary directories
and fixture metadata, never external live platforms as CI requirements. Real-media
checks are separate, authorized, optional smoke evidence. CI green cannot prove
every platform works. Accessibility automation catches some problems; manually
check keyboard order, announcements, zoom and reduced motion.

For exact commands see README. Each major stage updates VERIFICATION.md with
commands, outcomes and remaining gates. Do not replace an unrun check with a claim.
