# Public gallery evaluation

Date: 2026-09-29. Related issue: [#13](https://github.com/Lord-shaban/OpenDownload/issues/13).
Decision: [ADR 0003](adr/0003-public-gallery-adapter.md).

## Candidate and public access

Reviewed gallery-dl at upstream commit
[`8b9a56d6b1a53bc13a80894625fd58b24de3063a`](https://github.com/mikf/gallery-dl/commit/8b9a56d6b1a53bc13a80894625fd58b24de3063a).
No package was added and no gallery-dl extraction was run.

- [Instagram source](https://github.com/mikf/gallery-dl/blob/8b9a56d6b1a53bc13a80894625fd58b24de3063a/gallery_dl/extractor/instagram.py)
  contains post/sidecar parsing and a path without supplied login credentials.
  Login/challenge redirects abort extraction. This does not establish anonymous
  availability for a particular public post.
- [TikTok source](https://github.com/mikf/gallery-dl/blob/8b9a56d6b1a53bc13a80894625fd58b24de3063a/gallery_dl/extractor/tiktok.py)
  recognizes photo posts and emits image URLs. After invalid page data, its
  rehydration path calls `_solve_challenge`, even with zero retries.
  **Inference for OpenDownload:** integration needs an audited path that stops
  before this step to preserve our rule to terminate at access challenges.
- [CLI options](https://github.com/mikf/gallery-dl/blob/8b9a56d6b1a53bc13a80894625fd58b24de3063a/docs/options.md)
  include proxy, timeout, retry and default-config suppression (`--config-ignore`).
  Credentials, external extractors and command postprocessors are also available.
  A wrapper must provide fixed arguments and exclude those capabilities.

## Value and cost

Photo galleries are valuable and under-served by the current video/audio engine.
A second extractor shares our existing Python runtime but adds version pinning,
dependency auditing, a new output schema and access-policy review on upgrades.
The requester could not provide an authorized live photo corpus. Original test
images can verify our archive contract without proving platform reliability.

Retain direct images and bounded yt-dlp image entries. Defer a dedicated adapter
until the public-access and challenge-stop gates have evidence. This remains a
documented product gap.

## Coverage matrix

| Path                       | Evidence                                                          | Live platform status                              |
| -------------------------- | ----------------------------------------------------------------- | ------------------------------------------------- |
| Direct images              | Existing safe HTTP/MIME/byte tests                                | Live image smoke not run                          |
| Image-only entries → ZIP   | Generated PNGs and archive content tests                          | Depends on entries actually returned by yt-dlp    |
| Instagram-labeled contract | Original synthetic JSON, normalized choices and hidden asset URLs | No captured platform response or live photo smoke |
| TikTok-labeled contract    | Original synthetic JSON, normalized choices and hidden asset URLs | No captured platform response or live photo smoke |
| gallery-dl Instagram       | Pinned source review                                              | Deferred; anonymous extraction unverified         |
| gallery-dl TikTok          | Pinned source review identifies automatic challenge step          | Deferred; challenge-stop gate unmet               |

[Fixture provenance](../services/api/internal/media/testdata/gallery/README.md)
states that labels model our `RawInfo` contract, not platform wire schemas.

## Current safety evidence

`gallery_test.go` makes real HTTP requests through a local test proxy, compares
original PNG bytes after unzip and decodes them. It checks fixed names/order,
scratch removal, rejection of 21 images before HTTP, aggregate bytes against half
the job budget (reserving archive/scratch space), disguised HTML, 403 termination
and cancellation before the next image. The actual guarded proxy rejects
loopback/metadata assets. Existing tests cover DNS pinning and mixed/private DNS;
Linux CI separately verifies the container fence.

Workers re-analyze the source before download, enforce overall deadline/scratch
budget and remove files on failure. Asset URLs and output paths stay internal.
This change adds evidence without adding an extractor or runtime behavior.

## Integration gates

1. Obtain permission for single-image and multi-image posts on both platforms.
   Record permission, tool version, date and anonymous-access result. Do not
   commit cookies, signed URLs or unnecessary author data.
2. Capture sanitized actual extractor output and denied/login/challenge results.
   Prove challenges terminate, including the TikTok internal retry path.
3. Allow single-post image-only routing. Refuse profiles, mixed media, credential
   configuration, external extractors and command postprocessors.
4. Use fixed arguments/config, empty credential/cache state, bounded JSON and
   process lifetime. Route every request through the proxy/container fence and
   re-use 20-item, image-byte and job-byte limits.
5. Run live smoke separately from required deterministic CI and update the matrix.
   Stop on access denial; never switch credentials/access paths after denial.

Issue #13 stays open for authorized platform fixtures and integration gates.
Evaluation and local archive evidence are complete; live photo support is not.
