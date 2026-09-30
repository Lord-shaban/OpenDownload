#!/opt/engine/bin/python
"""Pinned anonymous YouTube attestation, inside the existing extractor fence."""
import os
import sys
import tempfile
from urllib.parse import urlsplit

YOUTUBE_HOSTS = frozenset((
    "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com",
    "youtu.be", "www.youtu.be", "youtube-nocookie.com", "www.youtube-nocookie.com",
))


def youtube_request(args):
    if len(args) < 2 or args[-2] != "--":
        return False
    try:
        url = urlsplit(args[-1])
        return (url.scheme in ("http", "https") and url.hostname in YOUTUBE_HOSTS
                and url.username is None and url.password is None)
    except ValueError:
        return False


def attestation_args(args):
    # Only this image-owned plugin directory is added after --no-plugin-dirs.
    # Do not enable default/user plugin discovery or accept a token from visitors.
    return [*args[:-2],
            "--plugin-dirs", "/opt/youtube-plugins",
            "--extractor-args", "youtube:player_client=mweb;fetch_pot=always",
            "--extractor-args", "youtubepot-bgutilscript:server_home=/opt/youtube-attestation/server",
            *args[-2:]]


def run(args, entrypoint):
    if not youtube_request(args):
        return entrypoint(args)
    # The provider's cache is separate from yt-dlp's --no-cache-dir. Bound its
    # lifetime to one extraction; remove it on success, refusal, and cancellation
    # that unwinds the interpreter. Container /tmp also bounds crash leftovers.
    previous = os.environ.get("XDG_CACHE_HOME")
    with tempfile.TemporaryDirectory(prefix="od-youtube-") as cache:
        os.environ["XDG_CACHE_HOME"] = cache
        try:
            return entrypoint(attestation_args(args))
        finally:
            if previous is None:
                os.environ.pop("XDG_CACHE_HOME", None)
            else:
                os.environ["XDG_CACHE_HOME"] = previous


if __name__ == "__main__":
    import yt_dlp
    run(sys.argv[1:], yt_dlp.main)
