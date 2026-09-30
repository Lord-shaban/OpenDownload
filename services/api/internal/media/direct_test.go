package media

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestActualGenericSingleFormatMetadata(t *testing.T) {
	// Captured shape of yt-dlp 2026.8.19 output for the owned sample MP4.
	const metadata = `{"title":"sample","direct":true,"url":"https://example.com/sample.mp4","ext":"mp4","extractor_key":"Generic","protocol":"https","format_id":"0","_type":"video"}`
	var raw RawInfo
	if err := json.Unmarshal([]byte(metadata), &raw); err != nil {
		t.Fatal(err)
	}
	analysis, err := Normalize(raw, raw.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Options) != 2 || analysis.Options[0].Kind != "video" || analysis.Options[0].Selector != "0" || analysis.Options[1].ID != "audio-mp3" || analysis.Options[1].Selector != "0" {
		t.Fatalf("top-level format lost: %+v", analysis.Options)
	}
	if analysis.Options[0].Label != "Original video" || analysis.Options[1].Detail != "Conversion · requires an audio track" {
		t.Fatal("invented source quality or audio availability")
	}
	raw.DRM = true
	if _, err := Normalize(raw, raw.URL); !errors.Is(err, ErrAccess) {
		t.Fatal("DRM direct file accepted")
	}
	raw.DRM = false
	raw.DetectedDRM = true
	if _, err := Normalize(raw, raw.URL); !errors.Is(err, ErrAccess) {
		t.Fatal("detected DRM direct file accepted")
	}
	raw.DetectedDRM = false
	for _, change := range []func(*RawInfo){
		func(r *RawInfo) { r.FormatID = "0;bad" },
		func(r *RawInfo) { r.Protocol = "file" },
		func(r *RawInfo) { r.Type = "url" },
		func(r *RawInfo) { r.Direct = false },
	} {
		candidate := raw
		change(&candidate)
		if _, err := Normalize(candidate, candidate.URL); !errors.Is(err, ErrUnsupported) {
			t.Fatal("unsafe or indirect root format accepted")
		}
	}
}
