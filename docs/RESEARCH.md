# Research ledger

Research date: 2026-09-29. Sources are primary project documentation. Versions are
resolved from registries at implementation time and committed in lockfiles.

| Source | Finding and decision |
|---|---|
| [Cobalt](https://github.com/imputnet/cobalt) | Fast paste-to-file UX and public-content boundary are good precedents. OpenDownload needs richer actual-format selection and temporary jobs. No source copied. |
| [MeTube](https://github.com/alexta69/metube) | Durable queue, concurrency, thumbnails/subtitles are valuable. Arbitrary yt-dlp overrides and cookies are inappropriate for this scope. No source copied. |
| [yt-dlp](https://github.com/yt-dlp/yt-dlp) | JSON metadata, explicit format IDs, native segmented download, progress templates, and FFmpeg postprocessing provide the engine. Support changes independently of our UI. Pin releases; do not self-update running workers. |
| [yt-dlp EJS](https://github.com/yt-dlp/yt-dlp/wiki/EJS) | Modern YouTube extraction needs an external JS runtime. Supply Node with bundled EJS; no remote script downloads. Access denials remain denials. |
| [Go net/http](https://pkg.go.dev/net/http#ServeMux) | Standard router supports method/path patterns and middleware composition. A framework provides no necessary feature for this API. |
| [SQLite WAL](https://www.sqlite.org/wal.html) | Good fit for one host and small, short write transactions. A single writer and bounded pools simplify contention. Not a shared database on network storage. |
| [modernc SQLite](https://gitlab.com/cznic/sqlite) | Pure Go avoids platform C toolchain requirements. The single backend library solves persistence portability. |
| [Next.js](https://nextjs.org/docs/app/getting-started/installation) | App Router, server-rendered shell, small interactive client island, standalone Docker output. Long media work belongs in Go. |
| [shadcn/ui](https://ui.shadcn.com/docs) | Source-owned primitives for focus, dialogs, buttons and controls; install only used components. |
| [OWASP SSRF](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html) | Initial URL checks do not cover redirects or rebinding. Validate every outbound destination and pin the resolved connection. Enforce network isolation. |
| [WCAG 2.2](https://www.w3.org/TR/WCAG22/) | Visible focus, labels, text contrast, reflow, statuses, keyboard access, and sufficient target sizes are design requirements. Automated checks are not certification. |
| [W3C motion](https://www.w3.org/WAI/WCAG22/Understanding/animation-from-interactions.html) | Respect reduced motion and avoid unnecessary motion. Use CSS state transitions; no animation package required. |
| [Chrome at I/O 2026](https://developer.chrome.com/blog/chrome-at-io26) | Modern transitions should preserve interactivity. Keep core behavior usable without new browser-only transition APIs. |
| [UI UX Pro Max](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill) | Installed and used before design. Flat utility style and loading/status guidance fit; its generated marketing layout does not fit the workspace. Explicit workspace override records the final direction. |

## Challenges to the original idea

"Universal" cannot mean every public URL is downloadable. yt-dlp's video coverage
does not imply image-gallery coverage, and upstream availability changes without
notice. Capabilities must be explicit. Generic direct images are useful now;
platform gallery extraction needs a dedicated adapter and fixtures before a claim.

The service cannot be both frictionless anonymous internet hosting and cheap,
unbounded media processing. Start with a local/trusted instance, small limits,
and no authentication import. Document requirements for public deployment.

Streaming directly from yt-dlp into the browser makes merges, retries, cancellation,
and browser refresh brittle. Process into a bounded temporary directory, then
stream completed output; explain the processing step in the UI.
