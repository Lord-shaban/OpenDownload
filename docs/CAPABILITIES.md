# Capabilities and honest platform support

Platform detection is a UI hint based on the hostname. The extractor is the
authority. yt-dlp covers YouTube, Instagram, TikTok, X, Facebook, Reddit, Vimeo,
SoundCloud, and many other public sources, subject to platform changes, region,
rate limits, and access policy. We do not certify every extractor or advertise
successful downloads for untested URLs.

The container installs pinned yt-dlp with its `curl-cffi` extra for extractors
that request browser-compatible TLS. Image builds verify Chrome support exists;
all resulting requests still pass through the guarded egress proxy. This adds
transport compatibility and does not remove a platform's login, human
verification, region or rate restrictions.

The cloud image also includes a pinned, image-owned anonymous YouTube PO token
provider for the public mobile web client. It supplies the platform's requested
player/media attestation; it does not import a visitor's account, cookies or
tokens, solve interactive CAPTCHA, enable DRM, or guarantee that a server IP
will be accepted. The Node provider runs only for YouTube inside the existing
process/network fence, requires the guarded loopback proxy, and keeps tokens in
memory for the extraction. There is no public token server or disk token cache.
YouTube's web requests use the installed Chrome TLS transport through that same
proxy. The provider's resolved dependencies are pinned with security fixes and
audited during image builds.
YouTube analysis and download share one extractor slot to bound Node memory;
waiting requests retain their original cancellation/timeout limits.
Upstream refusals still stop the request. See THIRD_PARTY_NOTICES.md for the
pinned source and license. Cloud verification of the reported URL is pending.

The next cloud candidate uses the upstream WPC provider with Chromium on a
private Xvfb display, to test the alternative of minting tokens in an actual
anonymous web client. The existing user/network/PID fence remains mandatory;
Chromium's inner setuid sandbox is unavailable under mapped UID 0. Browser
localhost bypasses are disabled, the CDP socket stays loopback, and the parent
cleans profiles when extraction stops. A private Linux container check verifies
that Chromium cannot navigate directly to the API through localhost. This
candidate is not yet proof that the reported YouTube URL works on the host.

Successful YouTube HTTP-format analysis is reused for two minutes in a bounded
in-memory snapshot cache (32 entries, at most 256 KiB each). Selected formats
are loaded through yt-dlp's stdin, without another browser launch or metadata
request. Only explicit media fields and non-credential headers are retained;
signed URLs never appear in API responses or a token file. Unsupported protocols
and subtitles keep their existing flow. Expired or refused media is not silently
re-extracted from the snapshot. This reduces repeated extraction; it does not
override upstream verification or rate limits.

On 2026-09-30, the reported YouTube URL `https://youtu.be/2cUkUbB3Gu4`
downloaded through the local real API without authentication but required human
verification from the blitz.cloud server. SoundCloud's public NASA Quindar sound
and direct Blender/GitHub video sources analyzed successfully on that host.
These outcomes describe the tested links, not platform-wide guarantees.
Issue #27 tracks the reported cloud YouTube/TikTok failures and subsequent
runtime/deployment verification.

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
