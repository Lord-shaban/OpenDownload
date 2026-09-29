package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/jobs"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
)

func TestFullFixtureFlowAndOwnership(t *testing.T) {
	dir := t.TempDir()
	store, err := jobs.Open(filepath.Join(dir, "jobs.db"), 5)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx, cancel := context.WithCancel(context.Background())
	m := &jobs.Manager{Store: store, Engine: media.Fixture{Delay: time.Millisecond}, Root: filepath.Join(dir, "files"), Workers: 1, MaxBytes: 1 << 20, Timeout: time.Second}
	if err := m.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); m.Wait() }()
	c := Config{Origin: "http://localhost:3000", Fixture: true, MaxBytes: 1 << 30, Retention: time.Hour}
	s := httptest.NewServer(New(c, m).Handler())
	defer s.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	call := func(method, path, body string) *http.Response {
		t.Helper()
		req, _ := http.NewRequest(method, s.URL+path, strings.NewReader(body))
		req.Header.Set("Origin", c.Origin)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}
	r := call("POST", "/api/v1/analyze", `{"url":"https://example.com/sample"}`)
	var a media.Analysis
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		t.Fatal(err)
	}
	_ = r.Body.Close()
	if r.StatusCode != 200 || a.ID == "" {
		t.Fatal("analysis failed")
	}
	r = call("POST", "/api/v1/jobs", `{"analysisId":"`+a.ID+`","optionId":"video-1080"}`)
	var j jobs.Job
	_ = json.NewDecoder(r.Body).Decode(&j)
	_ = r.Body.Close()
	if r.StatusCode != 202 {
		t.Fatal("create failed")
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		// HTTP JSON intentionally excludes owner and URL. Query via the session API.
		r = call("GET", "/api/v1/jobs", "")
		var listBody struct {
			Jobs []jobs.Job `json:"jobs"`
		}
		_ = json.NewDecoder(r.Body).Decode(&listBody)
		_ = r.Body.Close()
		if len(listBody.Jobs) > 0 {
			j = listBody.Jobs[0]
		}
		if j.State == "complete" {
			break
		}
	}
	if j.State != "complete" {
		t.Fatalf("job did not complete: %+v", j)
	}
	req, _ := http.NewRequest("GET", s.URL+"/api/v1/jobs/"+j.ID+"/files/0", nil)
	req.Header.Set("Range", "bytes=0-9")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 206 || len(data) != 10 || !strings.Contains(resp.Header.Get("Content-Disposition"), "attachment") {
		t.Fatal("range attachment failed")
	}
	other, _ := http.Get(s.URL + "/api/v1/jobs/" + j.ID)
	_ = other.Body.Close()
	if other.StatusCode != 404 {
		t.Fatal("ownership leak")
	}
	bad, _ := http.Post(s.URL+"/api/v1/analyze", "application/json", strings.NewReader(`{"url":"https://example.com"}`))
	_ = bad.Body.Close()
	if bad.StatusCode != 403 {
		t.Fatal("origin check missing")
	}
}
