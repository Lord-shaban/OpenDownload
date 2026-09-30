package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/jobs"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
)

type unavailableEngine struct {
	media.Fixture
	err error
}

func (e unavailableEngine) Analyze(context.Context, string) (media.Analysis, error) {
	return media.Analysis{}, fmt.Errorf("internal token=secret: %w", e.err)
}

func TestAnalysisReturnsSafeDistinctFailureCodes(t *testing.T) {
	for _, item := range []struct {
		err    error
		status int
		code   string
	}{
		{media.ErrPlatformVerification, 503, "platform_verification_required"},
		{media.ErrUpstreamForbidden, 503, "upstream_forbidden"},
		{media.ErrUpstreamRateLimit, 503, "upstream_rate_limited"},
		{media.ErrSourceConnection, 503, "source_connection_failed"},
		{media.ErrSourceMetadata, 503, "source_metadata_unavailable"},
		{media.ErrAccess, 422, "source_access_denied"},
	} {
		t.Run(item.code, func(t *testing.T) {
			s := New(Config{Origin: "https://download.example", Fixture: true}, &jobs.Manager{Engine: unavailableEngine{err: item.err}})
			r := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader(`{"url":"https://youtu.be/2cUkUbB3Gu4"}`))
			r.Header.Set("Origin", s.Config.Origin)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			var body struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if w.Code != item.status || body.Error.Code != item.code || body.Error.Message != item.err.Error() || strings.Contains(w.Body.String(), "secret") {
				t.Fatalf("incorrect public failure: %d %s", w.Code, w.Body.String())
			}
		})
	}
}
