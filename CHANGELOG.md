# Changelog

## [Unreleased]

### Added

- LinkedIn public video posts/activity URLs and `lnkd.in` resolution.
- Pinterest public video pins, original single-image pins, `pin.it` resolution
  and regional domains. Full boards and story galleries remain unsupported.
- Threads public post JSON extraction on `threads.com`/`threads.net`, including
  legacy `/t/` and redirecting share links. Supports a single video/image or
  at most 20 image-only carousel items using the existing guarded transport.
- Source logos/detection in both locales and a repeatable opt-in social smoke
  check for file bytes, media/ZIP validation, range serving and deletion.

### Fixed

- Preserve original progressive MP4 formats when LinkedIn/Pinterest/Threads
  omit codec metadata; rank unknown-resolution formats by reported bitrate.
  Do not invent resolution, codecs or confirmed audio-track availability.

These changes are in current source, not the immutable v0.1.0 images.

## [0.1.0] — 2026-10-01

First release of OpenDownload.

### Included

- Minimal glass workspace and `OpenDownload.` wordmark, light/dark themes,
  keyboard navigation, reduced motion and mobile layouts.
- English and Arabic, RTL and mixed-direction URL defenses.
- Public media analysis with actual available formats, video/audio downloads,
  MP3 conversion, direct images, thumbnails and available subtitle sidecars.
- Bounded extractor-provided image collections and ZIP output.
- SQLite job persistence, progress, cancellation, retry, owner-scoped files,
  HTTP ranges, expiry and deletion.
- Guarded outbound proxy, public-address/DNS checks, subprocess and storage
  bounds, non-root containers and a fail-closed cloud namespace profile.
- Public real-download deployment, protected PR delivery, required CI,
  versioned container publication with provenance/SBOM and post-publication checks.

### Release boundaries

- YouTube is unavailable in 0.1. Its investigation remains deferred.
- Dedicated Instagram/TikTok photo-gallery extraction is deferred. Existing
  image support does not imply universal photo-post support.
- Platform names are extractor candidates, not guarantees for every public URL.
- Private/paid content, login cookies, DRM, live streams, playlists, watermark
  removal and bulk downloading are outside scope.

[0.1.0]: https://github.com/Lord-shaban/OpenDownload/releases/tag/v0.1.0
