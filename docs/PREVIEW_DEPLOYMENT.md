# Cloud preview (fixture mode only)

This deployment demonstrates the interface and job/file flow. **It never downloads
real media.** The UI labels it as a test instance, and downloaded files explicitly
identify themselves as fixtures. Real downloads still require the isolated stack
in [SELF_HOSTING.md](SELF_HOSTING.md).

## Run locally

```sh
docker compose -f compose.preview.yaml up --build -d --wait
```

Open `http://localhost:3000`. Stop it with
`docker compose -f compose.preview.yaml down`; keep the named volume to preserve
fixture jobs and files. Files expire after 15 minutes.

## Container hosting

Build the repository root using `deploy/Dockerfile.preview`. Publish only port
3000, set `OD_ORIGIN` to the exact external HTTPS origin (without a trailing slash),
and mount persistent storage at `/data`. Leave the image's default start command.
The API runs inside this container on loopback from the web app's perspective;
no Compose DNS or separate API service is required.

The launcher forces fixture mode, one worker, a three-job queue, a 128 MiB per-job
budget, a five-minute deadline and 15-minute retention. Setting
`OD_FIXTURE_MODE=false` does not enable extraction. The image contains neither
yt-dlp nor FFmpeg. Both child processes stop when either fails, and SIGTERM is
forwarded so SQLite can close before the container exits.

For blitz.cloud, select only one running web part, choose this Dockerfile and
disable the previously detected API/egress parts. Check that `/data` is a kept
folder after deployment. The free plan can sleep; file backups require separate
arrangements. CI builds this image and verifies the fixture flow, persistence
through restart, and shutdown.

## Historical deployment attempts

The public site now runs the [real fenced cloud profile](FENCED_CLOUD.md), not
this preview. The preview remains an explicitly labeled demonstration for
development and testing. The observations below describe the earlier attempt.

The 2026-09-30 attempt to deploy the existing production layout on blitz.cloud
Free was rejected because it has two background parts; the account permits one.
A web + fixture API attempt built, but the platform did not resolve the Compose
hostname `api` (`getaddrinfo ENOTFOUND api`). This dedicated image fixes the preview
without removing the real extractor's network boundary.

The completed real deployment is recorded in [issue #22](https://github.com/Lord-shaban/OpenDownload/issues/22).
Hosting terms can change; consult [blitz.cloud's terms](https://blitz.cloud/terms/)
and [limits](https://blitz.cloud/docs/limits/). This preview does not establish
production compatibility or availability guarantees.
