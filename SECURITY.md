# Security policy

OpenDownload is pre-release. Only the latest development revision receives fixes.

Report vulnerabilities privately using GitHub's **Report a vulnerability** flow
when private vulnerability reporting is enabled for the repository. If unavailable,
contact the repository owner privately through their published GitHub contact
method. Never place exploit URLs, secrets, or sensitive logs in public issues.
No response-time guarantee is offered by this volunteer project.

Include affected revision, deployment mode, reproducible steps, impact, and a
minimal proof. Relevant issues include SSRF (including redirects and DNS rebinding),
job/file ownership violations, command injection, unsafe file paths, process escape,
and resource exhaustion. Upstream yt-dlp/FFmpeg vulnerabilities should also be
reported upstream.

Read `docs/THREAT_MODEL.md`. A hostname check by itself is not an SSRF defense.
The supported hardened topology prevents extractors from reaching the internet
except through the guarded proxy. Containers do not replace VM isolation for
hostile multi-tenant workloads. Public hosting requires additional abuse controls.
