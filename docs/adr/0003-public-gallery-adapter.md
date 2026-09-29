# ADR 0003: Defer a dedicated gallery extractor until public access is verified

Date: 2026-09-29. Status: accepted (defer integration).

## Context

Photo galleries have value beyond yt-dlp's primary video/audio coverage.
OpenDownload already archives bounded image-only entries and downloads direct
images. Another extractor is justified when it improves verified public coverage
within our access and resource boundaries.

Pinned gallery-dl review found photo parsing, but no authorized live corpus is
available. Its reviewed TikTok path solves a JavaScript challenge after invalid
page data; zero retries does not disable that step. See the
[evaluation, sources and matrix](../GALLERY_EVALUATION.md).

## Decision

Defer gallery-dl integration. Verify our archive contract with original synthetic
fixtures and generated PNGs. Keep #13 open for authorized platform fixtures and
an audited path that terminates at login/challenge/access denials. No credential
or challenge fallback is introduced.

## Consequences

Broader Instagram/TikTok photo support remains a documented gap. A later adapter
can reuse the media interface, guarded proxy, re-analysis, limits and ZIP writer
once its gates pass. Required CI stays deterministic; live platform availability
is recorded separately. We avoid the maintenance and access-policy cost of an
unverified runtime dependency.
