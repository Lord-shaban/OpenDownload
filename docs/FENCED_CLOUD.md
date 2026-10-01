# Real downloads in a single cloud container

This profile serves anonymous users and always runs the real extraction engine.
The original three-container [Compose deployment](SELF_HOSTING.md) and the
fixture-only preview are separate profiles.

## Required runtime boundary

`deploy/Dockerfile.cloud` runs the outer supervisor as UID/GID 10001. It starts
the API, Next.js, yt-dlp and FFmpeg inside an unprivileged Linux user/network
namespace with only loopback active and no external route, plus a PID namespace to
clean up descendants if the worker dies. Startup checks the namespace
identity, interfaces and a denied direct TCP connection. If namespace creation
or real dependency readiness fails, the public listener stays closed. There is
no fallback to fixture mode, a shared network or a privileged container.

Inactive, unaddressed Linux fallback devices `tunl0`, `sit0` and `ip6tnl0` are
permitted; every other extra interface or inspection failure is rejected.
Tini runs as PID 1 inside the worker namespace and reaps orphaned browser
descendants. The worker must be its direct child, PID 2.
The PID namespace controls process IDs and descendant cleanup. Docker's existing
`/proc` mount is retained to avoid remounting its masked paths; process metadata
for other components in the outer container can remain visible there. This
profile does not claim a private process-information filesystem for each job.

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
# Ubuntu with AppArmor 4: explicitly allow this container's nested namespaces.
sudo apparmor_parser -r deploy/apparmor-cloud
docker compose -f compose.fenced.yaml up --build -d --wait
```

This local profile drops capabilities, uses a read-only root, and permits nested
unprivileged namespaces in the Docker seccomp/AppArmor settings. The outer user
is still non-root. A production runtime must provide its own suitable sandbox
and allow namespace creation; startup refuses unsupported runtimes. Never use
`--privileged`, host networking or disable the namespace check to get it running.

The named AppArmor profile is a local/CI compatibility profile with `userns`
permission; it is not a full filesystem or syscall sandbox. It applies only to
this container and does not disable Ubuntu's system-wide namespace restrictions.
See [Ubuntu's AppArmor namespace guidance](https://documentation.ubuntu.com/security/security-features/privilege-restriction/apparmor/).
On other Linux distributions, use the runtime's equivalent namespace permission.

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
part. The complete image was deployed and verified on 2026-09-30 at
[opendownload.lord.blitz.cloud](https://opendownload.lord.blitz.cloud/), initially
at commit `da66d4e276f92c79937b58ba2992424c4e59aec2`. Its readiness endpoint
reported real mode, FFmpeg and yt-dlp available. The owned two-second sample
produced a byte-identical 25,532-byte MP4 and a decodable 50,302-byte MP3.
Live checks also verified cancellation, owner isolation, HTTP ranges, private
target refusal and expired media returning 404 after the 15-minute retention.
An image downloaded through the Arabic UI survived a deployment restart.
Issue #22 records subsequent deployment commits and verification evidence.
Free apps sleep after two hours without visitors and wake on the next
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
