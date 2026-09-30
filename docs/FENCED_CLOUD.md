# Real downloads in a single cloud container

This profile serves anonymous users and always runs the real extraction engine.
The original three-container [Compose deployment](SELF_HOSTING.md) and the
fixture-only preview are separate profiles.

## Required runtime boundary

`deploy/Dockerfile.cloud` runs the outer supervisor as UID/GID 10001. It starts
the API, Next.js, yt-dlp and FFmpeg inside an unprivileged Linux user/network
namespace with only loopback and no external route, plus a PID namespace to
clean up descendants if the worker dies. Startup checks the namespace
identity, interfaces and a denied direct TCP connection. If namespace creation
or real dependency readiness fails, the public listener stays closed. There is
no fallback to fixture mode, a shared network or a privileged container.

Private Unix sockets in a mode-0700 directory carry:

- Web ingress to Next.js; only port 3000 is published.
- HTTP proxy traffic to the existing guarded egress policy. Every destination
  is validated and dialed at a pinned public IP, including extractor requests.
- Initial URL DNS validation to a broker that rejects private or mixed DNS
  answers. The isolated API never falls back to host DNS.
- A private isolation diagnostic used by container checks; it is not web routed.

The process supervisor terminates API/web when either fails. Process groups
handle normal cancellation and shutdown; broker/bridge failures stop the profile.
This shares a container and filesystem
between application components; it does not claim a separate VM per download or
protection from a complete application/container compromise. Do not mount host
credentials, Docker sockets or unrelated data. See [THREAT_MODEL.md](THREAT_MODEL.md).

## Conservative public limits

The cloud launcher forces real mode and these settings even if old preview
environment variables are present:

| Setting                       | Value      |
| ----------------------------- | ---------- |
| Processing workers            | 1          |
| Queued/active jobs            | 3          |
| Maximum job scratch/output    | 128 MiB    |
| Stored-media admission budget | 512 MiB    |
| Processing timeout            | 5 minutes  |
| File retention                | 15 minutes |

The admission budget counts retained media plus a whole-job reservation before
processing. The existing scratch watchdog cancels jobs over their job limit.
These are application limits, not filesystem quotas: writes can overshoot
between watchdog ticks and the SQLite database is outside the media budget.
Leave disk headroom and monitor usage. Per-session ownership, queue controls,
analysis concurrency, same-origin mutations and the existing global mutation
rate limit remain enforced. Anonymous users share finite capacity; a successful
sample does not establish throughput at arbitrary public traffic levels.

## Local Linux verification

```sh
docker compose -f compose.fenced.yaml up --build -d --wait
```

This local profile drops capabilities, uses a read-only root, and permits nested
unprivileged namespaces in the Docker seccomp/AppArmor settings. The outer user
is still non-root. A production runtime must provide its own suitable sandbox
and allow namespace creation; startup refuses unsupported runtimes. Never use
`--privileged`, host networking or disable the namespace check to get it running.

CI builds the actual image and runs `scripts/cloud-check.py`: real owned test
video, MP3 conversion, exact video bytes, private target refusal, denied direct
egress, range streaming, owner isolation, persistent media across restart,
fail-closed startup and graceful shutdown. The sample is the repository's own
test pattern, served from the exact tested commit on GitHub. No social platform
uptime is used as an integration gate.

## blitz.cloud setup

Use `deploy/Dockerfile.cloud`, leave the start command empty to use its image
command, publish only the web part on port 3000, and keep API/egress parts off.
Keep `/data` with sufficient storage, and set `OD_ORIGIN` to the exact HTTPS
origin. No passwords are required for visitors. Provider account credentials
must not be passed into this container.

The Free account inspected on 2026-09-30 provides five apps, 512 MB reserved
memory and 10 GB storage, with one background part. This profile uses one web
part. The provider's sandbox allowed a preliminary user/network namespace probe;
the complete image still requires deployment verification before being called
live. Free apps sleep after two hours without visitors and wake on the next
visit. Free stored files do not receive the Pro nightly file backup. No paid
plan, keep-awake option, trial or custom-domain purchase is required.

References: [runtime sandbox](https://blitz.cloud/docs/deploy-docker-image/),
[limits](https://blitz.cloud/docs/limits/).

## Rollback and operations

Deploy only a commit with successful required GitHub Actions checks. Record the
live commit, real media results and restart evidence in issue #22. To roll back,
choose the prior verified commit/profile, rebuild, and retain `/data`. Returning
to `Dockerfile.preview` restores fixture mode and is not a real download service.
Run one API process against its SQLite volume; do not share it across replicas.
Monitor readiness, storage and worker failures. Platform access restrictions,
DRM or required login remain unsupported; no access bypass is included.
