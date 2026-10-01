package media

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
)

const maxYouTubeSnapshot = 256 << 10

type youtubeSnapshot struct {
	source   string
	analysis Analysis
}

func (e *YTDLP) cachedYouTubeSnapshot(source string) (Analysis, bool) {
	e.snapshotMu.Lock()
	defer e.snapshotMu.Unlock()
	fresh := make([]youtubeSnapshot, 0, len(e.snapshots))
	var result Analysis
	found := false
	for _, entry := range e.snapshots {
		if !time.Now().Before(entry.analysis.snapshotExpires) {
			continue
		}
		fresh = append(fresh, entry)
		if entry.source == source {
			result, found = entry.analysis, true
			result.Options = append([]Option(nil), result.Options...)
		}
	}
	e.snapshots = fresh
	return result, found
}

func (e *YTDLP) rememberYouTubeSnapshot(source string, analysis Analysis) {
	e.snapshotMu.Lock()
	defer e.snapshotMu.Unlock()
	// At most 8 MiB, independent of the public analysis/queue capacities.
	if len(e.snapshots) >= 32 {
		e.snapshots = e.snapshots[1:]
	}
	e.snapshots = append(e.snapshots, youtubeSnapshot{source, analysis})
}

func snapshotHeaders(raw json.RawMessage) map[string]string {
	var input map[string]string
	_ = json.Unmarshal(raw, &input)
	result := map[string]string{}
	for key, value := range input {
		key = http.CanonicalHeaderKey(key)
		if len(value) > 2048 || strings.ContainsAny(value, "\r\n") {
			continue
		}
		switch key {
		case "User-Agent", "Accept", "Accept-Language", "Origin", "Referer":
			result[key] = value
		}
	}
	return result
}

func makeYouTubeSnapshot(data []byte, info RawInfo, analysis Analysis) []byte {
	var raw struct {
		ID      string                       `json:"id"`
		Headers json.RawMessage              `json:"http_headers"`
		Formats []map[string]json.RawMessage `json:"formats"`
	}
	if json.Unmarshal(data, &raw) != nil || !safeID.MatchString(raw.ID) || len(raw.Formats) != len(info.Formats) {
		return nil
	}
	formats := []map[string]any{}
	for i, format := range info.Formats {
		// Cache complete HTTP files only. Other protocols/subtitles retain their
		// existing extraction flow rather than losing required manifest fields.
		if !safeFormat(format) || format.Protocol != "https" && format.Protocol != "http" {
			continue
		}
		if _, err := security.Parse(format.URL); err != nil {
			continue
		}
		formats = append(formats, map[string]any{
			"format_id": format.ID, "ext": format.Ext, "url": format.URL,
			"protocol": format.Protocol, "vcodec": format.VCodec, "acodec": format.ACodec,
			"height": format.Height, "width": format.Width, "fps": format.FPS,
			"filesize": format.Bytes, "filesize_approx": format.Approx,
			"http_headers": snapshotHeaders(raw.Formats[i]["http_headers"]),
		})
	}
	if len(formats) == 0 {
		return nil
	}
	// Explicit fields exclude cookies, paths, commands, playlists and webpage_url
	// (yt-dlp would otherwise retry extraction automatically after a media refusal).
	snapshot, err := json.Marshal(map[string]any{
		"_type": "video", "id": raw.ID, "title": analysis.Title,
		"duration": analysis.Duration, "extractor_key": "Youtube",
		"formats": formats, "http_headers": snapshotHeaders(raw.Headers),
	})
	if err != nil || len(snapshot) > maxYouTubeSnapshot {
		return nil
	}
	return snapshot
}

func snapshotSupports(analysis Analysis, option Option) bool {
	if analysis.snapshotSource != analysis.URL || !time.Now().Before(analysis.snapshotExpires) || len(analysis.extractorSnapshot) == 0 || len(analysis.extractorSnapshot) > maxYouTubeSnapshot || option.Kind != "video" && option.Kind != "audio" {
		return false
	}
	var info struct {
		Formats []RawFormat `json:"formats"`
	}
	if json.Unmarshal(analysis.extractorSnapshot, &info) != nil {
		return false
	}
	ids := map[string]bool{}
	for _, format := range info.Formats {
		ids[format.ID] = true
	}
	for _, id := range strings.Split(option.Selector, "+") {
		if !ids[id] {
			return false
		}
	}
	return true
}
