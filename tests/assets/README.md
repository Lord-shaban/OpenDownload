# Real media integration sample

`sample.mp4` is an original two-second test pattern and sine tone generated for
OpenDownload. It is covered by the repository's MIT license. No third-party
footage or music is included.

Generation command (FFmpeg 9.0.2):

```sh
ffmpeg -f lavfi -i 'testsrc=size=160x90:rate=12' \
  -f lavfi -i 'sine=frequency=440:sample_rate=44100' -t 2 \
  -c:v libx264 -pix_fmt yuv420p -c:a aac -movflags +faststart sample.mp4
```

The cloud container check downloads the committed public file through the real
yt-dlp engine and guarded proxy, checks its bytes, and converts it to MP3 using
FFmpeg. This does not use the fixture engine or depend on a social platform's
availability.
