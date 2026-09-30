#!/usr/bin/python3
"""Image-owned Chromium entrypoint; all page requests use the guarded proxy."""
import os
import sys


def chromium_args(args):
    # Preserve normal site/process separation and restrict the debugging socket.
    # The browser driver is a local trusted client; web origins get no CDP grant.
    filtered = [a for a in args if not a.startswith((
        "--proxy-", "--remote-allow-origins", "--remote-debugging-host=",
        "--remote-debugging-address=", "--disable-features=",
    )) and a != "--disable-site-isolation-trials"]
    return [*filtered,
            "--proxy-server=http://127.0.0.1:8090",
            "--proxy-bypass-list=<-loopback>",
            "--remote-debugging-address=127.0.0.1",
            "--site-per-process", "--disable-background-networking",
            "--disable-breakpad", "--disable-crash-reporter", "--disable-gpu",
            "--disable-dev-shm-usage", "--disable-quic", "--no-sandbox",
            "--renderer-process-limit=2", "--num-raster-threads=1",
            "--autoplay-policy=document-user-activation-required"]


if __name__ == "__main__":
    # Inside the unprivileged worker's user/PID/network namespace. Chromium sees
    # mapped UID 0; its setuid sandbox is unavailable. The outer fence is required.
    os.execv("/usr/lib/chromium/chromium", ["chromium", *chromium_args(sys.argv[1:])])
