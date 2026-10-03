# LinkedIn, Pinterest and Threads

These adapters are included in current source after v0.1.0. Build from source
with `docker compose up --build`; the v0.1.0 tag and published container images
retain their original scope. Updating the product website does not deploy the API.

| Source | Accepted links | Media and exclusions |
| --- | --- | --- |
| LinkedIn | Public `/posts/...`, `/feed/update/urn:li:activity:...`, redirecting `lnkd.in` | Public video through pinned yt-dlp; source captions/thumbnail when available. Learning courses, private/account-only content excluded. |
| Pinterest | Individual `/pin/<id>/` or slugged pins on upstream-supported regional domains, redirecting `pin.it` | Public video through yt-dlp; original single-image pins through the anonymous Pin resource. Boards, story galleries and account-only content excluded. |
| Threads | `/@user/post/<code>`, legacy `/t/<code>`, redirecting `/share/...` on `threads.com` or `threads.net` | Single video/image or image-only carousels of at most 20 items. Mixed/multiple-video carousels, text-only posts, private profiles and unavailable metadata excluded. |

## Extraction and network boundary

LinkedIn and Pinterest use the existing pinned yt-dlp extractors. Some progressive
files have neither codec nor resolution metadata. The normalizer preserves those
original files, ranks missing-resolution files by reported bitrate and reports
only known dimensions. MP3 conversion is conditional on an actual audio track.

Pinterest's image adapter requests the same anonymous Pin endpoint used upstream,
checks the requested pin ID and requires the original image to be on `pinimg.com`.
Video/story results and unavailable image metadata delegate to the existing extractor.
It never substitutes a video thumbnail for a single-image pin.

Threads has no built-in extractor in the pinned yt-dlp build. Its public
search-preview representation contains post JSON. The adapter requests that
representation with the upstream-style
`Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)` User-Agent.
This is an explicit compatibility header, not a verified crawler identity or
an availability guarantee. It uses no account/cookies, does not solve challenges,
and stops on denied/private responses or missing requested-post metadata.
Changes to that representation can break extraction.

HTML is tokenized and JSON parsed structurally, never evaluated as JavaScript.
The adapter matches the exact shortcode, so recommendations/replies cannot
substitute for the requested post. CDN URLs remain absent from public API JSON.
Workers reanalyze on processing/retry; format IDs use dimensions so reordering
the source's video versions does not change the chosen quality.

All native metadata uses the existing operator-configured guarded HTTP proxy,
including redirects. Reads are capped at 8 MiB, redirects at five, and requests
respect cancellation/timeouts. Media passes through the same guarded proxy.
The existing image sniffing, aggregate gallery budget, job limits, output checks,
range serving, ownership, expiry and deletion remain in force.

## Verification

`social_test.go` covers source identity, unknown codecs, bitrate ranking, HTTP
errors, metadata limits, private/text/mixed posts, image selection, URL privacy,
regional domains and short-link destination boundaries. Frontend unit and browser
tests cover detection, logos and the bilingual list.

Use the opt-in smoke check against a real instance and a permitted public sample:

```sh
python scripts/social-smoke.py --source '<public-post-url>' --kind video
python scripts/social-smoke.py --source '<public-image-post-url>' --kind image
```

Video/image files are verified with FFprobe; image ZIPs are checked for valid
bounded image entries. The check validates actual file bytes, HTTP ranges and
delete revocation, then removes the remote job. It saves local evidence under
`.data/social-smoke-output`. `--base-url`, `--origin`, `--ffprobe` and `--output`
configure the instance and tools. Samples are opt-in, never platform-uptime CI gates.

Exact dated results belong to [VERIFICATION.md](VERIFICATION.md), not a universal
platform-support claim. Existing v0.1 release evidence remains historical.
