# Product brief

## The opportunity

Downloaders often hide the available file choices behind presets, ask users to
interpret codec jargon, or obscure long-running work. OpenDownload should make
the actual capabilities of a URL understandable and keep users in control of
processing and temporary retention.

Primary audience: individuals and small trusted groups saving their own or
permitted public media on a self-hosted instance. Anonymous public hosting is
not the MVP deployment target. No promise of universal platform availability.

## Essential MVP

One URL → detect source locally → analyze on explicit submission → display
metadata and actual options → choose → queue → show progress → download.
Include cancellation, retry, job history, expiry, delete, accessible errors,
direct images, thumbnails, and available subtitles. Offer audio conversion and
video qualities with codecs/container labels. Preserve the original where useful.

Missing requirements resolved:

- Formats come from analysis, never from a hard-coded quality menu.
- Source format and final container are separate concepts; conversion costs time.
- Analysis tokens expire; stale signed media URLs must be re-extracted at download.
- Work survives refresh. Process interruptions are shown as retryable failures.
- Metadata, URL history, and files have bounded retention. No external analytics.
- Files stream with byte-range support; they are never read into application memory.
- Capabilities are determined per item. A famous platform name is not a promise.
- Public-only means a login/access denial is a terminal error, not a reason to
  inject cookies or alternate clients.
- No full channels, playlists, live streams, automatic subscriptions, or bulk jobs.

## Product evolution

| Idea                                                  | Value           | Complexity / maintenance            | Performance / security                         | Decision                         |
| ----------------------------------------------------- | --------------- | ----------------------------------- | ---------------------------------------------- | -------------------------------- |
| More complete gallery adapter                         | High            | Medium, separate upstream extractor | Public access/challenge-stop evidence required | Deferred; see gallery evaluation |
| Locale/RTL support                                    | High            | Medium                              | Bidi URL defenses and production browser tests | English/Arabic implemented       |
| Filename preferences                                  | Medium          | Low                                 | Allow safe preset names only                   | Next                             |
| Remember local quality preferences                    | Medium          | Low                                 | Avoid storing submitted URLs in browser        | Next                             |
| Resumable interrupted processing                      | Medium          | High                                | Partial file validation, leases                | Later                            |
| Playlist preview + explicit bounded selection         | High            | High                                | Abuse and disk amplification                   | Later                            |
| Optional object storage / distributed workers         | Medium          | High                                | Useful only beyond a single node               | Later                            |
| Browser extension                                     | Medium          | Medium                              | Extra access surface                           | Later                            |
| Cookies / private downloads / DRM / watermark removal | Outside purpose | High                                | Access-control and policy costs                | Excluded                         |

Success criteria: new users complete an authorized download without reading docs;
errors explain the next action; cancellation terminates work and frees capacity;
temporary files disappear on schedule; selection never claims unavailable quality.
Measure locally via tests and structured operational counters, not user tracking.
