# Capabilities — OpenDownload 0.1

## Platform scope

**YouTube is not supported in v0.1.** The API returns `youtube_unavailable`
(HTTP 422) before source DNS/extraction. It also rejects YouTube extractor results
and persisted YouTube processing/retries. The experimental attestation/browser
work in [PR #30](https://github.com/Lord-shaban/OpenDownload/pull/30) is not part
of this release. Its investigation is deferred, not fixed.

The compact sites list includes Instagram, TikTok, X/Twitter, Facebook, Reddit,
Vimeo and SoundCloud. These are candidate extractors: availability depends on
each URL, region, upstream changes and the host's access. The list does not
certify every extractor or imply that every public post is downloadable.

| Source                                | Actual evidence / release limit                                                                                                     |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| TikTok video                          | The reported public video downloaded as a 2,178,061-byte MP4 with audio/video, 30.600 s, from the real cloud instance on 2026-10-01 |
| SoundCloud                            | The public NASA Quindar sample downloaded as a 296,596-byte M4A, 24.288 s, from the real cloud instance                             |
| Direct video/audio                    | Real owned GitHub test pattern and Blender trailer; video/MP3, ranges, ownership, deletion and persistence checks                   |
| Direct image                          | Owned public JPEG downloaded with exact bytes, ranges and delete revocation                                                         |
| Instagram, X, Facebook, Reddit, Vimeo | Upstream adapters available; no release-wide live URL matrix or availability guarantee                                              |
| Social photo galleries                | Dedicated Instagram/TikTok gallery adapter deferred; only safe image entries already returned by the extractor                      |
| YouTube                               | Explicitly unavailable in 0.1                                                                                                       |

Exact tagged-image and release-deployment evidence belongs to
[release issue #36](https://github.com/Lord-shaban/OpenDownload/issues/36).
Earlier samples establish only those tested links, not universal platform support.

The container installs pinned yt-dlp with `curl-cffi` for extractors that request
browser-compatible TLS. Every request still passes through the guarded egress
proxy. Transport compatibility does not override login, human verification,
region, rate or media access restrictions. Release images include no experimental
YouTube browser/token provider, account cookies or public token service.

## Media

| Media     | Adapter                                | Scope                                                         |
| --------- | -------------------------------------- | ------------------------------------------------------------- |
| Video     | yt-dlp + FFmpeg local merge            | Actually offered qualities, compatible container; no DRM/live |
| Audio     | yt-dlp, optional FFmpeg MP3 conversion | Source audio and MP3; no fake lossless conversion             |
| Thumbnail | Guarded HTTP download                  | Extractor-provided public JPEG/PNG/WebP                       |
| Subtitles | yt-dlp                                 | Available manual/automatic languages, VTT sidecar             |
| Image     | Guarded HTTP download                  | Public direct JPEG/PNG/WebP/GIF; safe Content-Type detection  |
| Gallery   | Extractor image entries                | At most 20 images, bounded aggregate storage, ZIP output      |

yt-dlp is primarily a video/audio extractor. Many social photo galleries are not
returned as image collections. A dedicated adapter remains deferred after
[gallery-dl evaluation](GALLERY_EVALUATION.md); public access and challenge-stop
evidence are required before integration. This is a documented future feature.

Quality varies by URL. Video-only streams require available audio. Unsupported
network protocols are excluded. WebM/MKV may be offered; MP4 only appears with
compatible codecs. Thumbnail/subtitle choices appear only when analysis reports
them. Original watermarks stay intact. Private/paid/authenticated media, DRM,
live streams, full playlists, bulk jobs and access bypass are outside scope.
