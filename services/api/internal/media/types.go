package media

import (
	"context"
	"time"
)

type Option struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Label     string `json:"label"`
	Extension string `json:"extension"`
	Detail    string `json:"detail"`
	Bytes     int64  `json:"bytes,omitempty"`
	// Internal selections never appear in API output or accept arbitrary client flags.
	Selector  string   `json:"-"`
	AssetURL  string   `json:"-"`
	AssetURLs []string `json:"-"`
	Language  string   `json:"-"`
	Automatic bool     `json:"-"`
}

type Analysis struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Creator      string   `json:"creator"`
	Platform     string   `json:"platform"`
	Duration     float64  `json:"duration"`
	Options      []Option `json:"options"`
	ExpiresAt    string   `json:"expiresAt"`
	URL          string   `json:"-"`
	Thumbnail    string   `json:"-"`
	HasThumbnail bool     `json:"hasThumbnail"`
	Owner        string   `json:"-"`
	// Server-owned, short-lived metadata; never serialized or accepted from clients.
	extractorSnapshot []byte
	snapshotSource    string
	snapshotExpires   time.Time
}

type Progress struct {
	Percent float64
	Phase   string
}
type Output struct {
	Name  string
	Path  string
	Bytes int64
}

type Engine interface {
	Analyze(context.Context, string) (Analysis, error)
	Download(context.Context, Analysis, Option, string, func(Progress)) ([]Output, error)
}
