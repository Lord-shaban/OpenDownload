package media

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var ErrUnsupported = errors.New("this public link does not offer supported downloadable media")
var ErrAccess = errors.New("private, authenticated, live, or DRM-protected content is not supported")
var safeID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
var safeLanguage = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

type RawFormat struct {
	ID       string  `json:"format_id"`
	Ext      string  `json:"ext"`
	Height   int     `json:"height"`
	Width    int     `json:"width"`
	FPS      float64 `json:"fps"`
	VCodec   string  `json:"vcodec"`
	ACodec   string  `json:"acodec"`
	Protocol string  `json:"protocol"`
	Bytes    int64   `json:"filesize"`
	Approx   int64   `json:"filesize_approx"`
	DRM      bool    `json:"has_drm"`
	URL      string  `json:"url"`
}
type RawSubtitle struct {
	Ext string `json:"ext"`
	URL string `json:"url"`
}
type RawInfo struct {
	Title        string                   `json:"title"`
	Uploader     string                   `json:"uploader"`
	Extractor    string                   `json:"extractor_key"`
	Duration     float64                  `json:"duration"`
	Thumbnail    string                   `json:"thumbnail"`
	Availability string                   `json:"availability"`
	Live         bool                     `json:"is_live"`
	DRM          bool                     `json:"has_drm"`
	Type         string                   `json:"_type"`
	URL          string                   `json:"url"`
	Ext          string                   `json:"ext"`
	Formats      []RawFormat              `json:"formats"`
	Subtitles    map[string][]RawSubtitle `json:"subtitles"`
	Automatic    map[string][]RawSubtitle `json:"automatic_captions"`
	Entries      []RawInfo                `json:"entries"`
}

func Normalize(raw RawInfo, source string) (Analysis, error) {
	if raw.DRM || raw.Live || raw.Availability == "private" || raw.Availability == "premium_only" || raw.Availability == "subscriber_only" || raw.Availability == "needs_auth" {
		return Analysis{}, ErrAccess
	}
	a := Analysis{Title: truncate(raw.Title, 300), Creator: truncate(raw.Uploader, 120), Platform: truncate(raw.Extractor, 60), Duration: raw.Duration, URL: source, Thumbnail: raw.Thumbnail, Options: []Option{}}
	if a.Title == "" {
		a.Title = "Untitled media"
	}
	a.HasThumbnail = raw.Thumbnail != ""
	if len(raw.Entries) > 0 {
		if len(raw.Entries) > 20 {
			return Analysis{}, errors.New("collections are limited to 20 images")
		}
		images := []string{}
		for _, entry := range raw.Entries {
			if entry.DRM || entry.Live || !imageExt(entry.Ext) || entry.URL == "" {
				return Analysis{}, errors.New("video playlists are outside this release; paste a single item")
			}
			images = append(images, entry.URL)
		}
		a.Options = append(a.Options, Option{ID: "gallery", Kind: "image", Label: fmt.Sprintf("%d original images", len(images)), Extension: "zip", Detail: "Bounded image collection", AssetURLs: images})
		return a, nil
	}
	if imageExt(raw.Ext) && raw.URL != "" {
		a.Options = append(a.Options, Option{ID: "image", Kind: "image", Label: "Original image", Extension: raw.Ext, Detail: "Source file, unchanged", AssetURL: raw.URL})
	}
	formats := append([]RawFormat(nil), raw.Formats...)
	sort.SliceStable(formats, func(i, j int) bool {
		if formats[i].Height == formats[j].Height {
			return formats[i].FPS > formats[j].FPS
		}
		return formats[i].Height > formats[j].Height
	})
	var audio *RawFormat
	for i := range formats {
		f := &formats[i]
		if safeFormat(*f) && f.VCodec == "none" && f.ACodec != "none" && f.ACodec != "" {
			if audio == nil || (f.Ext == "m4a" && audio.Ext != "m4a") || f.Bytes > audio.Bytes && f.Ext == audio.Ext {
				audio = f
			}
		}
	}
	seen := map[string]bool{}
	for _, f := range formats {
		if !safeFormat(f) || f.VCodec == "none" || f.VCodec == "" || f.Height <= 0 {
			continue
		}
		ext := f.Ext
		selector := f.ID
		detail := fmt.Sprintf("%s · %s", strings.ToUpper(ext), truncate(f.VCodec, 24))
		bytes := f.Bytes
		if bytes == 0 {
			bytes = f.Approx
		}
		if f.ACodec == "none" || f.ACodec == "" {
			if audio == nil {
				continue
			}
			selector += "+" + audio.ID
			if !(ext == "mp4" && audio.Ext == "m4a") && !(ext == "webm" && audio.Ext == "webm") {
				ext = "mkv"
			}
			detail = fmt.Sprintf("%s · audio included", strings.ToUpper(ext))
			bytes += audio.Bytes
		} else {
			detail += " · audio included"
		}
		key := fmt.Sprintf("%d-%s", f.Height, ext)
		if seen[key] {
			continue
		}
		seen[key] = true
		a.Options = append(a.Options, Option{ID: "video-" + f.ID, Kind: "video", Label: fmt.Sprintf("%dp", f.Height), Extension: ext, Detail: detail, Selector: selector, Bytes: bytes})
		if len(a.Options) >= 16 {
			break
		}
	}
	if audio != nil {
		a.Options = append(a.Options, Option{ID: "audio-source", Kind: "audio", Label: "Original audio", Extension: audio.Ext, Detail: "Source audio, no conversion", Selector: audio.ID, Bytes: audio.Bytes})
		a.Options = append(a.Options, Option{ID: "audio-mp3", Kind: "audio", Label: "MP3 audio", Extension: "mp3", Detail: "Converted · up to 192 kbps", Selector: audio.ID})
	} else {
		for _, f := range formats {
			if safeFormat(f) && f.ACodec != "none" && f.ACodec != "" {
				a.Options = append(a.Options, Option{ID: "audio-mp3", Kind: "audio", Label: "MP3 audio", Extension: "mp3", Detail: "Converted from source · up to 192 kbps", Selector: f.ID})
				break
			}
		}
	}
	if raw.Thumbnail != "" {
		a.Options = append(a.Options, Option{ID: "thumbnail", Kind: "image", Label: "Thumbnail", Extension: "image", Detail: "Original cover image", AssetURL: raw.Thumbnail})
	}
	for _, group := range []struct {
		data map[string][]RawSubtitle
		auto bool
	}{{raw.Subtitles, false}, {raw.Automatic, true}} {
		langs := []string{}
		for lang := range group.data {
			langs = append(langs, lang)
		}
		sort.Strings(langs)
		count := 0
		for _, lang := range langs {
			if !safeLanguage.MatchString(lang) {
				continue
			}
			for _, sub := range group.data[lang] {
				if sub.Ext == "vtt" {
					id := "subtitle-" + lang
					detail := "Original captions · WebVTT"
					if group.auto {
						id = "automatic-" + lang
						detail = "Auto-generated · WebVTT"
					}
					a.Options = append(a.Options, Option{ID: id, Kind: "subtitle", Label: lang, Extension: "vtt", Detail: detail, Language: lang, Automatic: group.auto})
					count++
					break
				}
			}
			if count >= 20 {
				break
			}
		}
	}
	if len(a.Options) == 0 {
		return Analysis{}, ErrUnsupported
	}
	return a, nil
}

func safeFormat(f RawFormat) bool {
	if f.DRM || !safeID.MatchString(f.ID) {
		return false
	}
	switch f.Protocol {
	case "https", "http", "http_dash_segments", "m3u8_native", "m3u8":
	default:
		return false
	}
	switch f.Ext {
	case "mp4", "webm", "mkv", "m4a", "mp3", "ogg", "opus", "flac", "wav":
		return true
	}
	return false
}
func imageExt(ext string) bool {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg", "png", "webp", "gif":
		return true
	}
	return false
}
func truncate(value string, max int) string {
	r := []rune(strings.TrimSpace(value))
	if len(r) > max {
		r = r[:max]
	}
	return string(r)
}
