# Changelog

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
