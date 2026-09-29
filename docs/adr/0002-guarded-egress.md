# ADR 0002: Extractors use a guarded outbound proxy

Status: accepted. Date: 2026-09-29.

Validate initial URLs and every HTTP proxy request/HTTPS CONNECT target. Reject
local/reserved IPs and DNS sets containing any unsafe address. Pin the selected
public IP at dial time. Allow standard HTTP(S) ports only. Containers deny direct
egress from API/extractors; their public traffic must pass through the proxy.
The web container uses a separate bridge for its published loopback port and is
not inside that egress fence. It proxies a fixed API address and does not fetch
submitted media URLs or run extraction. See the [Compose verification evidence](../VERIFICATION.md).

Why: Go URL validation cannot constrain yt-dlp's subsequent network requests,
redirects, media manifests, or DNS rebinding. A proxy alone also cannot protect
against extractor code that ignores it, so network isolation is part of the design.

FFmpeg handles local postprocessing only. Disable config/plugin/credential imports.
Proxy enforces the destination address, not the ethics of content. Public-only
policy is also applied at extraction normalization and exposed in the UI.

Consequence: some unusual ports/protocols/extractors are unsupported. Native
development is less isolated and must not be advertised as a hardened deployment.
