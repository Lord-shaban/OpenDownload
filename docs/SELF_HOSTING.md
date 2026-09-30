# Self-hosting and operations

The initial deployment is a single host for an individual or trusted small group.
Do not run multiple APIs on one data directory or expose it as an anonymous public
downloader without further abuse protection and stronger worker sandboxing.

For a cloud demo that never extracts real media, see
[the fixture-only preview deployment](PREVIEW_DEPLOYMENT.md).

## Configuration

Native Go processes read environment variables. Compose reads `.env` and provides
explicit values to containers. The web app uses server-only `API_URL`; no backend
key is sent to the browser.

| Variable | Default | Purpose |
|---|---|---|
| OD_PORT | 8080 | API port |
| OD_DATA_DIR | .data | SQLite and generated per-job files |
| OD_EGRESS_PROXY | http://127.0.0.1:8090 | Mandatory proxy for real media extraction |
| OD_ORIGIN | http://localhost:3000 | Exact permitted browser origin |
| OD_WORKERS | 2 | Concurrent processing jobs (1–8) |
| OD_QUEUE_LIMIT | 20 | Global active+queued job limit |
| OD_MAX_BYTES | 536870912 | Maximum scratch/output bytes per job (512 MiB) |
| OD_JOB_TIMEOUT | 15m | Wall-clock processing deadline |
| OD_RETENTION | 1h | Job/file retention from creation |
| OD_FIXTURE_MODE | false | Explicit test engine, visibly labeled in UI |
| OD_YTDLP | yt-dlp | Operator-controlled executable path |
| API_URL | http://127.0.0.1:8080 | Next.js same-origin API rewrite destination |

## Storage and cleanup

Job state and files share a persistent local volume. Put it on a filesystem with
an enforced disk quota. Application byte limits are a watchdog and may overshoot
briefly; they do not replace a kernel quota. Keep free disk space well above
worker count × per-job budget. The periodic sweeper removes expired jobs/files;
startup also clears abandoned scratch directories. No permanent media library.

Back up SQLite using SQLite's backup API or stop the API before copying the data
volume. Do not copy only the database file while WAL writes are active. Restores
recover queued jobs; processing jobs are marked interrupted. Temporary media can
be discarded and should not be included in a durable backup by default.

## Network and security

Compose only publishes `127.0.0.1:3000`. API containers live on an internal network.
Egress proxy bridges that network to public internet and validates destinations.
It has no published host port. Proxy ports allow HTTP 80 and HTTPS 443 only.
The web container uses a separate ingress bridge for its published loopback port;
it is not subject to the extractor's egress fence. It proxies a fixed API address
and never fetches submitted source URLs or runs extraction. Making every web
network internal also prevents host port publication on the tested Docker engine.
Do not add another external network to API, host-network it, or bypass the proxy.

Trusted native development lacks a network fence; environment proxies are not an
isolation boundary. No cookies, credentials, proxies from users or arbitrary flags.
Update the pinned yt-dlp and distribution-provided FFmpeg through reviewed image rebuilds; no live
self-updating workers. YouTube support may fail due to platform access restrictions;
do not fix that by importing credentials or evading denied access.

## Public deployment checklist

Use TLS, a reverse proxy with identity-aware access, per-user quotas and bandwidth
limits, host disk quotas, VM/worker isolation, monitoring, a vulnerability upgrade
process, and local legal/access-policy review. Forwarded client IP headers are not
trusted by default. Session capability tokens scope jobs but are not full account
authentication. Set the exact external OD_ORIGIN and secure HTTPS cookie behavior.

## Operations

`GET /api/v1/health`: liveness. `GET /api/v1/status`: dependency readiness, fixture
mode and configured limits. Logs are JSON with request ID, method, route status
and duration; no submitted URLs or raw extractor stderr. Use request IDs in bug
reports. Reverse proxy logs must also avoid source URL/body logging.

For a temporary operator investigation, `OD_DIAGNOSTIC_PUBLIC_KEY` accepts a
base64 SPKI RSA public key (2048-4096 bits). Failed extractions then log an
AES-256-GCM envelope with an RSA-OAEP-SHA256 wrapped key and the last 8 KiB of
stderr. Both encryption layers use `OpenDownload diagnostic v1` as associated
data/label. Keep the private key on the operator's device, never on the server;
the visitor response remains canonical. This mode defaults off. Remove the
public-key setting and restart after collecting the needed evidence. Decrypted
diagnostics may contain signed media URLs/tokens and must not be posted in
issues or stored in Git.

SIGTERM stops accepting new work, cancels active process groups, and waits for
workers before closing SQLite. Expiry is visible in the UI. Access-denied,
unsupported and resource-limit errors are user-actionable; do not endlessly retry.
