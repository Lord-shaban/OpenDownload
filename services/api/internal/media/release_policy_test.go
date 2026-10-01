package media

import (
	"context"
	"errors"
	"testing"
)

func TestReleaseExcludesYouTubeBeforeRunningTools(t *testing.T) {
	e := &YTDLP{Binary: "must-not-run"}
	for _, source := range []string{"https://youtu.be/2cUkUbB3Gu4", "https://www.youtube.com/watch?v=public", "https://m.youtube.com/shorts/public", "https://www.youtube-nocookie.com/embed/public", "https://YouTube.com./watch?v=public"} {
		if !YouTubeSource(source) {
			t.Fatal("YouTube hostname was not recognized", source)
		}
		if _, err := e.Analyze(context.Background(), source); !errors.Is(err, ErrYouTubeUnavailable) {
			t.Fatal("analysis invoked tools or returned the wrong policy", err)
		}
		if _, err := e.Download(context.Background(), Analysis{URL: source}, Option{}, t.TempDir(), func(Progress) {}); !errors.Is(err, ErrYouTubeUnavailable) {
			t.Fatal("persisted download bypassed release policy", err)
		}
	}
	for _, source := range []string{"https://youtube.com.attacker.net/a", "https://notyoutube.com/a", "https://example.com/youtube.com", "https://www.tiktok.com/@creator/video/1"} {
		if YouTubeSource(source) {
			t.Fatal("unrelated host was excluded", source)
		}
	}
}

func TestIndirectYouTubeMetadataExcluded(t *testing.T) {
	if _, err := Normalize(RawInfo{Extractor: "Youtube"}, "https://example.com/redirect"); !errors.Is(err, ErrYouTubeUnavailable) {
		t.Fatal("indirect YouTube result bypassed release scope", err)
	}
}
