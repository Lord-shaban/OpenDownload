#!/opt/engine/bin/python
"""Pinned anonymous YouTube attestation, inside the existing extractor fence."""
import sys
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
            "--impersonate", "chrome",
            "--extractor-args", "youtube:player_client=mweb;fetch_pot=always",
            "--extractor-args", "youtubepot-bgutilscript:server_home=/opt/youtube-attestation/server/runtime",
            *args[-2:]]


def run(args, entrypoint):
    if not youtube_request(args):
        return entrypoint(args)
    return entrypoint(attestation_args(args))


if __name__ == "__main__":
    import yt_dlp
    run(sys.argv[1:], yt_dlp.main)
