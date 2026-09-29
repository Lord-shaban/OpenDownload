package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/jobs"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
)

func TestAnalysisOwnershipExpiryAndPayloadLimits(t *testing.T) {
	s := New(Config{Origin: "http://localhost:3000", Fixture: true}, &jobs.Manager{Engine: media.Fixture{}})
	handler := s.Handler()
	call := func(path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.Header.Set("Origin", s.Config.Origin)
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	w := call("/api/v1/analyze", `{"url":"https://example.com/sample"}`, nil)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var analysis media.Analysis
	if err := json.Unmarshal(w.Body.Bytes(), &analysis); err != nil {
		t.Fatal(err)
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("session cookie is not protected")
	}
	create := `{"analysisId":"` + analysis.ID + `","optionId":"video-1080"}`
	if w := call("/api/v1/jobs", create, nil); w.Code != 410 {
		t.Fatal("another session used the analysis")
	}
	s.mu.Lock()
	entry := s.cache[analysis.ID]
	entry.expires = time.Now().Add(-time.Second)
	s.cache[analysis.ID] = entry
	s.mu.Unlock()
	if w := call("/api/v1/jobs", create, cookie); w.Code != 410 {
		t.Fatal("expired analysis was used")
	}
	for _, body := range []string{`{"url":"https://example.com","cookies":"secret"}`, `{"url":"https://example.com"}{}`, `{"url":"https://example.com/` + strings.Repeat("a", 9000) + `"}`} {
		if w := call("/api/v1/analyze", body, cookie); w.Code != 400 {
			t.Fatalf("bad payload returned %d", w.Code)
		}
	}
	for i := 0; i < cap(s.analyses); i++ {
		s.analyses <- struct{}{}
	}
	if w := call("/api/v1/analyze", `{"url":"https://example.com"}`, cookie); w.Code != 503 {
		t.Fatal("concurrent analysis limit ignored")
	}
}
func TestMutationRateLimit(t *testing.T) {
	s := New(Config{Origin: "http://localhost:3000", Fixture: true}, &jobs.Manager{Engine: media.Fixture{}})
	handler := s.Handler()
	for i := 0; i < 61; i++ {
		r := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader(`{"url":"https://example.com/unsupported"}`))
		r.Header.Set("Origin", s.Config.Origin)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		want := 422
		if i == 60 {
			want = 429
		}
		if w.Code != want {
			t.Fatalf("request %d returned %d", i, w.Code)
		}
		if i == 60 && w.Header().Get("Retry-After") != "60" {
			t.Fatal("missing retry hint")
		}
	}
}
