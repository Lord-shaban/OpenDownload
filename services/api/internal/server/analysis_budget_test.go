package server

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/jobs"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
)

type pendingAnalysis struct {
	media.Fixture
	entered chan time.Time
}

func (e pendingAnalysis) Analyze(ctx context.Context, _ string) (media.Analysis, error) {
	deadline, _ := ctx.Deadline()
	e.entered <- deadline
	<-ctx.Done()
	return media.Analysis{}, ctx.Err()
}

func TestAnalysisBudgetIsBoundedAndCallerCanCancel(t *testing.T) {
	for _, item := range []struct {
		url    string
		budget time.Duration
	}{
		{"https://youtu.be/2cUkUbB3Gu4", 90 * time.Second},
		{"https://www.youtube.com/watch?v=2cUkUbB3Gu4", 90 * time.Second},
		{"https://youtube.com.evil.example/video", 45 * time.Second},
		{"https://www.tiktok.com/@example/video/123", 45 * time.Second},
	} {
		t.Run(item.url, func(t *testing.T) {
			engine := pendingAnalysis{entered: make(chan time.Time, 1)}
			s := New(Config{Origin: "https://download.example", Fixture: true}, &jobs.Manager{Engine: engine})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r := httptest.NewRequest("POST", "/api/v1/analyze", strings.NewReader(`{"url":"`+item.url+`"}`)).WithContext(ctx)
			r.Header.Set("Origin", s.Config.Origin)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			started := time.Now()
			done := make(chan struct{})
			go func() {
				s.Handler().ServeHTTP(w, r)
				close(done)
			}()
			select {
			case deadline := <-engine.entered:
				if duration := deadline.Sub(started); duration < item.budget-time.Second || duration > item.budget+time.Second {
					t.Errorf("analysis budget outside limit: %v", duration)
				}
			case <-time.After(2 * time.Second):
				cancel()
				t.Fatal("analysis never started")
			}
			cancel()
			select {
			case <-done:
				if w.Code != 504 || !strings.Contains(w.Body.String(), "analysis_timeout") {
					t.Fatalf("canceled analysis: %d %s", w.Code, w.Body.String())
				}
			case <-time.After(2 * time.Second):
				t.Fatal("caller cancellation did not stop analysis")
			}
		})
	}
}
