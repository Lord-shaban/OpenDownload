package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestYouTubeExcludedBeforeDNSOrEngineInBothModes(t *testing.T) {
	for _, fixture := range []bool{false, true} {
		s := New(Config{Origin: "https://download.example", Fixture: fixture}, nil)
		for _, source := range []string{"https://youtu.be/2cUkUbB3Gu4", "https://www.youtube.com/watch?v=public", "https://www.youtube-nocookie.com/embed/public"} {
			r := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader(`{"url":"`+source+`"}`))
			r.Header.Set("Origin", s.Config.Origin)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			if w.Code != 422 || !strings.Contains(w.Body.String(), `"code":"youtube_unavailable"`) {
				t.Fatalf("release policy was not returned: %d %s", w.Code, w.Body.String())
			}
		}
	}
}
