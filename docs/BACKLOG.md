# Granular backlog

## Release boundary

The first release is v0.1.0. M0–M4, Arabic/RTL, workspace refinement and real
cloud hosting are accepted. Release delivery is tracked in
[OD-036 / #36](https://github.com/Lord-shaban/OpenDownload/issues/36).
YouTube (#27) and dedicated social photo-gallery extraction (#13) are deferred
beyond 0.1, not claimed complete. The historical implementation slices follow.

Canonical issue seeds are in `scripts/github-plan.json`. `scripts/github-setup.py`
creates missing labels, milestones, and issues idempotently; it never marks work
complete merely because code exists. Actual remote issue numbers are recorded by
the script. This file explains the small reviewable implementation slices.

| Key    | Milestone | Work                                 | Acceptance                                                                                                        |
| ------ | --------- | ------------------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| OD-001 | M0        | Research and architecture            | Sources, tradeoffs, access boundary, ADRs committed                                                               |
| OD-002 | M0        | Design system and repository hygiene | Tokens, contributor/security policy, templates, CI                                                                |
| OD-003 | M1        | URL and outbound security            | Reserved IPs, mixed DNS, pinned dial, redirects tests                                                             |
| OD-004 | M1        | Metadata and format normalization    | Safe video/audio/subtitle/thumbnail/image options, no signed URLs                                                 |
| OD-005 | M1        | Analysis limits and sessions         | TTL, bounds, owner scope, timeouts, rate limits                                                                   |
| OD-006 | M2        | SQLite queue and restart recovery    | Atomic claim, bounded capacity, interrupted-state recovery                                                        |
| OD-007 | M2        | Worker lifecycle                     | Progress, deadlines, cancellation, retries, descendant cleanup                                                    |
| OD-008 | M2        | Temporary storage and range serving  | Safe paths, expiry/delete, owner checks, range response                                                           |
| OD-009 | M3        | Paste/analyze/select interaction     | Accessible feedback, capability-driven choices, keyboard flow                                                     |
| OD-010 | M3        | Jobs and responsive workspace        | Queue/retry/cancel/download, mobile and reduced motion                                                            |
| OD-011 | M4        | Docker and operations                | Isolated egress, limits, health, graceful shutdown, backup                                                        |
| OD-012 | M4        | Release verification                 | Deterministic E2E + real authorized smoke + isolation evidence                                                    |
| OD-013 | M5        | Gallery adapter evaluation           | Document actual gallery-dl public-only behavior and cost                                                          |
| OD-014 | M5        | Locale and RTL groundwork            | Locale messages and bidi-safe URL controls                                                                        |
| OD-015 | M0        | Enforce GitHub delivery workflow     | Protected main, required CI, versioned policy, accurate linked verification                                       |
| OD-016 | M6        | Redesign the media workspace         | Subtle glass, wordmark, clearer flows, themes, RTL and production browser gates                                   |
| OD-022 | M7        | Public cloud deployment              | HTTPS, persistent data, guarded real downloads and live permitted media                                           |
| OD-036 | 0.1.0     | First release                        | YouTube excluded, consistent docs/version, required CI, published-image tests, live deployment and GitHub release |

Release blockers: no false platform/gallery support promises; no cookie/DRM
paths; all safe selection values must come from a verified analysis; no unbounded
subprocess output or filesystem growth; container isolation validated on Linux.
