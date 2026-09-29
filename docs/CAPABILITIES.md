# Capabilities and honest platform support

Platform detection is a UI hint based on the hostname. The extractor is the
authority. yt-dlp covers YouTube, Instagram, TikTok, X, Facebook, Reddit, Vimeo,
SoundCloud, and many other public sources, subject to platform changes, region,
rate limits, and access policy. We do not certify every extractor or advertise
successful downloads for untested URLs.

| Media     | Initial adapter                        | Scope                                                           |
| --------- | -------------------------------------- | --------------------------------------------------------------- |
| Video     | yt-dlp + FFmpeg local merge            | Actual offered qualities, compatible container; no DRM/live     |
| Audio     | yt-dlp, optional FFmpeg MP3 conversion | Source audio and MP3; no fake lossless conversion               |
| Thumbnail | Guarded HTTP download                  | Extractor-provided public JPEG/PNG/WebP                         |
| Subtitles | yt-dlp                                 | Actual available manual/automatic languages, VTT sidecar        |
| Image     | Guarded HTTP download                  | Public direct JPEG/PNG/WebP/GIF; detect using safe Content-Type |
| Gallery   | Extractor image entries                | Small bounded image-only collections if yt-dlp returns them     |

**Limit:** yt-dlp is primarily a video/audio extractor. Many Instagram and TikTok
photo galleries will not be returned as image collections. A dedicated adapter
is deferred after [gallery-dl evaluation](GALLERY_EVALUATION.md); anonymous access
and challenge termination need evidence before integration. Its addition must
preserve public-only policy, SSRF controls and per-gallery bounds.
This is a deliberate, documented gap against the broader product vision.

Quality availability varies by URL. Video-only formats must be merged with an
available audio stream. Formats with unsupported network protocols are excluded.
Video may arrive as WebM or MKV; MP4 is offered only where the codecs are compatible.
Subtitles/thumbnail choices appear only when analysis reports them. Selecting a
subtitle language never invokes authentication. Original watermarks are preserved.
