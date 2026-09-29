# Threat model

Assets: host/network, instance bandwidth/disk/CPU, submitted URLs, job metadata,
and temporary files. Attackers control submitted URLs and upstream media responses.
Extractor and codec parsing run outside the request handler and are untrusted.

| Threat                               | Controls                                                                                                                                                 | Residual risk                                                                                                              |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| SSRF / redirects / DNS rebinding     | HTTP(S) only, restricted ports, no credentials, every proxy destination checked, public addresses only, resolved IP pinned, API network internal         | Native development has no network fence; never deploy that topology to untrusted users                                     |
| IPv4/IPv6 metadata and local ranges  | Explicit reserved prefixes, reject mixed public/private DNS, normalize mapped IPv4                                                                       | New special-use allocations require policy updates                                                                         |
| Command injection                    | exec argument arrays, `--ignore-config`, no plugins, no arbitrary flags, source URL after `--`, safe format ID validation                                | Upstream vulnerabilities require timely pinned-version upgrades                                                            |
| Process leaks / cancel races         | Context deadline, Linux process-group kill, bounded waits, terminal transition guards                                                                    | Windows development process-tree behavior differs                                                                          |
| Disk / CPU / bandwidth exhaustion    | Worker limit, queue/per-owner limits, job deadline, output/scratch byte watchdog, container memory/CPU/PID limits, small gallery limit                   | Local volumes still need host disk quota/monitoring; application watchdog overshoot is bounded in time, not a kernel quota |
| Cross-user file/history access       | Random HttpOnly session, hashed owner binding, every analysis/job/file route scoped                                                                      | Trusted self-host sessions are not full account authentication                                                             |
| CSRF / origin spoofing               | Exact allowed origins for mutations, JSON only, SameSite cookie, no open CORS                                                                            | Non-browser clients can obtain sessions; public abuse controls still required                                              |
| File path traversal / symlinks       | Generated directory IDs, fixed templates, resolved relative path checks, reject symlinks and nonregular files                                            | Container compromise is outside application path guarantees                                                                |
| Upstream secrets / privacy leaks     | No cookies, no raw extraction JSON or signed URLs returned, redacted logs, bounded retention, no analytics                                               | Upstream sees instance IP and requested content                                                                            |
| DRM / private / paywall / watermarks | Reject DRM/live/auth-only metadata, no credential API, no DRM enablement or filter pipeline                                                              | Extraction policy is conservative and platforms can mislabel availability                                                  |
| Malicious images / subtitles         | Download as attachment, no raw HTML rendering, proxy validated images, fixed acceptable extensions                                                       | Client applications must safely parse downloaded media                                                                     |
| Bidirectional text spoofing          | Reject raw/percent-encoded bidi controls in submitted URLs; strip formatting controls from displayed titles; isolate URL/codec/filename fields in RTL UI | Ordinary Arabic text is accepted; third-party metadata remains untrusted                                                   |

## Deployment boundary

Compose uses an **internal** API network. API/yt-dlp/FFmpeg cannot make direct
internet connections. The proxy alone bridges to the outbound network. The proxy
has no host-published port. No Docker socket, host networking, privileged mode,
or host home-directory mount. Linux containers run as a dedicated unprivileged
user with dropped capabilities, no-new-privileges, read-only root and writable
scratch/data volumes only. FFmpeg consumes local files only.

Production checks include private destination URLs, mixed DNS, mapped IPv6,
redirect to private target, direct extractor egress denial, cancellation of
FFmpeg descendants, disk pressure, and concurrent delete/serve behavior.
See verification for which checks have actually run. Container configuration
alone is not evidence of successful isolation.
