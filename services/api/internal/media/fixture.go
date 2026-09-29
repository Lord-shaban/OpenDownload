package media

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Fixture is an explicit test engine. It never contacts a platform or pretends
// its demonstration output is a real media download.
type Fixture struct{ Delay time.Duration }

func (f Fixture) Analyze(ctx context.Context, source string) (Analysis, error) {
	select {
	case <-ctx.Done():
		return Analysis{}, ctx.Err()
	case <-time.After(f.Delay):
	}
	if strings.Contains(source, "unsupported") {
		return Analysis{}, ErrUnsupported
	}
	return Analysis{Title: "A small film about the great outdoors", Creator: "OpenDownload sample", Platform: "Fixture", Duration: 154, URL: source, Options: []Option{
		{ID: "video-1080", Kind: "video", Label: "1080p", Extension: "mp4", Detail: "MP4 · audio included", Bytes: 28400000, Selector: "1080"},
		{ID: "video-720", Kind: "video", Label: "720p", Extension: "mp4", Detail: "MP4 · audio included", Bytes: 14200000, Selector: "720"},
		{ID: "audio-mp3", Kind: "audio", Label: "MP3 audio", Extension: "mp3", Detail: "Converted · up to 192 kbps", Selector: "audio"},
		{ID: "subtitle-en", Kind: "subtitle", Label: "English", Extension: "vtt", Detail: "Original captions · WebVTT", Language: "en"},
	}}, nil
}
func (f Fixture) Download(ctx context.Context, a Analysis, opt Option, dir string, update func(Progress)) ([]Output, error) {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(f.Delay):
		}
		update(Progress{Percent: float64(i * 10), Phase: "downloading"})
	}
	if strings.Contains(a.URL, "fail") {
		return nil, errors.New("fixture: demonstration failure; retry or try a different link")
	}
	update(Progress{Percent: 99, Phase: "processing"})
	name := "fixture-readme.txt"
	data := []byte("OpenDownload test fixture. This file verifies the download flow. It is not extracted media.\n")
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, err
	}
	return []Output{{Name: name, Path: path, Bytes: int64(len(data))}}, nil
}
