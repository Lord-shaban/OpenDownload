package media

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeSelections(t *testing.T) {
	raw := RawInfo{Title: "Example", Thumbnail: "https://cdn.example/secret.jpg", Formats: []RawFormat{
		{ID: "137", Ext: "mp4", Height: 1080, VCodec: "avc1", ACodec: "none", Protocol: "https"},
		{ID: "140", Ext: "m4a", VCodec: "none", ACodec: "aac", Protocol: "https"},
		{ID: "evil;rm", Ext: "mp4", Height: 2160, VCodec: "avc1", ACodec: "aac", Protocol: "https"},
		{ID: "drm", Ext: "mp4", Height: 2160, VCodec: "avc1", ACodec: "aac", Protocol: "https", DRM: true},
	}}
	a, err := Normalize(raw, "https://example.com/source")
	if err != nil {
		t.Fatal(err)
	}
	if a.Options[0].Selector != "137+140" || a.Options[0].Extension != "mp4" {
		t.Fatalf("bad merge: %+v", a.Options[0])
	}
	b, _ := json.Marshal(a)
	if strings.Contains(string(b), "secret.jpg") || strings.Contains(string(b), "137+140") || strings.Contains(string(b), "https://example.com/source") {
		t.Fatal("internal selection leaked")
	}
	if len(a.Options) != 4 {
		t.Fatalf("unsafe formats surfaced: %d", len(a.Options))
	}
}
func TestAccessAndPlaylists(t *testing.T) {
	for _, raw := range []RawInfo{{DRM: true}, {Live: true}, {Availability: "needs_auth"}, {Entries: []RawInfo{{Ext: "mp4", URL: "https://example.com/v"}}}} {
		if _, err := Normalize(raw, "https://example.com"); err == nil {
			t.Fatal("accepted access-restricted media or video playlist")
		}
	}
}
