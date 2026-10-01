package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/jobs"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/media"
	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
)

type cacheEntry struct {
	analysis media.Analysis
	expires  time.Time
}
type rateEntry struct {
	count int
	since time.Time
}
type responseLog struct {
	http.ResponseWriter
	status int
}

func (w *responseLog) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}
func (w *responseLog) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
func (w *responseLog) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type Server struct {
	Config   Config
	Manager  *jobs.Manager
	Policy   security.Policy
	mu       sync.Mutex
	cache    map[string]cacheEntry
	rates    map[string]rateEntry
	analyses chan struct{}
}

func New(config Config, manager *jobs.Manager) *Server {
	s := &Server{Config: config, Manager: manager, cache: map[string]cacheEntry{}, rates: map[string]rateEntry{}, analyses: make(chan struct{}, 4)}
	if config.ResolverSocket != "" {
		s.Policy.Resolver = security.NewSocketResolver(config.ResolverSocket)
	}
	return s
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /api/v1/status", s.status)
	mux.HandleFunc("POST /api/v1/analyze", s.analyze)
	mux.HandleFunc("GET /api/v1/analyses/{id}/thumbnail", s.thumbnail)
	mux.HandleFunc("POST /api/v1/jobs", s.create)
	mux.HandleFunc("GET /api/v1/jobs", s.list)
	mux.HandleFunc("GET /api/v1/jobs/{id}", s.get)
	mux.HandleFunc("POST /api/v1/jobs/{id}/cancel", s.cancel)
	mux.HandleFunc("POST /api/v1/jobs/{id}/retry", s.retry)
	mux.HandleFunc("DELETE /api/v1/jobs/{id}", s.delete)
	mux.HandleFunc("GET /api/v1/jobs/{id}/files/{file}", s.file)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := jobs.ID()
		logged := &responseLog{ResponseWriter: w}
		w = logged
		defer func() {
			status := logged.status
			if status == 0 {
				status = http.StatusOK
			}
			slog.Info("request", "request_id", requestID, "method", r.Method, "route", r.Pattern, "status", status, "duration_ms", time.Since(start).Milliseconds())
		}()
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.Method != "GET" && r.Method != "HEAD" {
			if r.Header.Get("Origin") != s.Config.Origin {
				problem(w, 403, "origin_denied", "This request did not come from the configured OpenDownload origin.")
				return
			}
			if r.Method != "DELETE" {
				kind, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if kind != "application/json" {
					problem(w, 415, "json_required", "Send a JSON request.")
					return
				}
			}
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			if !s.allow(host) {
				w.Header().Set("Retry-After", "60")
				problem(w, 429, "rate_limited", "Too many requests. Try again in a minute.")
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}
func (s *Server) allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, v := range s.rates {
		if now.Sub(v.since) > time.Minute {
			delete(s.rates, k)
		}
	}
	entry := s.rates[key]
	if entry.since.IsZero() {
		if len(s.rates) >= 1000 {
			return false
		}
		entry.since = now
	}
	entry.count++
	s.rates[key] = entry
	return entry.count <= 60
}
func (s *Server) owner(w http.ResponseWriter, r *http.Request) string {
	raw := ""
	if cookie, err := r.Cookie("od_session"); err == nil {
		if decoded, err := hex.DecodeString(cookie.Value); err == nil && len(decoded) == 32 {
			raw = cookie.Value
		}
	}
	if raw == "" {
		raw = jobs.ID() + jobs.ID()
		http.SetCookie(w, &http.Cookie{Name: "od_session", Value: raw, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(s.Config.Origin, "https://"), SameSite: http.SameSiteStrictMode, MaxAge: 86400})
	}
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	_ = s.owner(w, r)
	_, ytErr := exec.LookPath(s.Config.Binary)
	_, ffErr := exec.LookPath("ffmpeg")
	ready := s.Config.Fixture || ytErr == nil && ffErr == nil
	writeJSON(w, 200, map[string]any{"ready": ready, "fixtureMode": s.Config.Fixture, "version": "0.1.0", "limits": map[string]any{"workers": s.Config.Workers, "queue": s.Config.Queue, "maxBytes": s.Config.MaxBytes, "retentionSeconds": int(s.Config.Retention.Seconds())}, "dependencies": map[string]bool{"ytDlp": ytErr == nil, "ffmpeg": ffErr == nil}})
}
func decode(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("extra data")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func problem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func (s *Server) cached(id, owner string) (media.Analysis, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[id]
	if !ok || entry.analysis.Owner != owner || time.Now().After(entry.expires) {
		return media.Analysis{}, false
	}
	return entry.analysis, true
}
func (s *Server) analyze(w http.ResponseWriter, r *http.Request) {
	owner := s.owner(w, r)
	var input struct {
		URL string `json:"url"`
	}
	if decode(w, r, &input) != nil {
		problem(w, 400, "invalid_request", "Enter one public URL.")
		return
	}
	if u, err := security.Parse(input.URL); err == nil && media.YouTubeSource(u.String()) {
		problem(w, 422, "youtube_unavailable", media.ErrYouTubeUnavailable.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	var normalized string
	if s.Config.Fixture {
		u, err := security.Parse(input.URL)
		if err != nil {
			problem(w, 400, "unsafe_url", err.Error())
			return
		}
		normalized = u.String()
	} else {
		u, err := s.Policy.Validate(ctx, input.URL)
		if err != nil {
			problem(w, 400, "unsafe_url", security.ErrUnsafeURL.Error())
			return
		}
		normalized = u.String()
	}
	select {
	case s.analyses <- struct{}{}:
		defer func() { <-s.analyses }()
	default:
		problem(w, 503, "analysis_busy", "Analysis is busy. Try again in a moment.")
		return
	}
	a, err := s.Manager.Engine.Analyze(ctx, normalized)
	if err != nil {
		if ctx.Err() != nil {
			problem(w, 504, "analysis_timeout", "Analysis took too long or was canceled. Try a shorter public link.")
			return
		}
		switch {
		case errors.Is(err, media.ErrYouTubeUnavailable):
			problem(w, 422, "youtube_unavailable", media.ErrYouTubeUnavailable.Error())
		case errors.Is(err, media.ErrPlatformVerification):
			problem(w, 503, "platform_verification_required", media.ErrPlatformVerification.Error())
		case errors.Is(err, media.ErrUpstreamForbidden):
			problem(w, 503, "upstream_forbidden", media.ErrUpstreamForbidden.Error())
		case errors.Is(err, media.ErrUpstreamRateLimit):
			problem(w, 503, "upstream_rate_limited", media.ErrUpstreamRateLimit.Error())
		case errors.Is(err, media.ErrSourceConnection):
			problem(w, 503, "source_connection_failed", media.ErrSourceConnection.Error())
		case errors.Is(err, media.ErrSourceMetadata):
			problem(w, 503, "source_metadata_unavailable", media.ErrSourceMetadata.Error())
		case errors.Is(err, media.ErrAccess):
			problem(w, 422, "source_access_denied", media.ErrAccess.Error())
		default:
			problem(w, 422, "source_unavailable", err.Error())
		}
		return
	}
	a.ID = jobs.ID()
	a.Owner = owner
	expires := time.Now().Add(10 * time.Minute)
	a.ExpiresAt = expires.UTC().Format(time.RFC3339)
	s.mu.Lock()
	for id, entry := range s.cache {
		if time.Now().After(entry.expires) {
			delete(s.cache, id)
		}
	}
	if len(s.cache) >= 200 {
		s.mu.Unlock()
		problem(w, 503, "analysis_busy", "Analysis capacity reached. Try again shortly.")
		return
	}
	s.cache[a.ID] = cacheEntry{a, expires}
	s.mu.Unlock()
	writeJSON(w, 200, a)
}
func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	owner := s.owner(w, r)
	var input struct {
		AnalysisID string `json:"analysisId"`
		OptionID   string `json:"optionId"`
	}
	if decode(w, r, &input) != nil {
		problem(w, 400, "invalid_request", "Select an available format.")
		return
	}
	a, ok := s.cached(input.AnalysisID, owner)
	if !ok {
		problem(w, 410, "analysis_expired", "This analysis expired. Analyze the link again.")
		return
	}
	for _, option := range a.Options {
		if option.ID == input.OptionID {
			if option.Bytes > s.Config.MaxBytes {
				problem(w, 413, "file_too_large", "This source exceeds the configured file limit. Choose a smaller format.")
				return
			}
			now := time.Now().UTC()
			j, err := s.Manager.Store.Create(jobs.Job{Owner: owner, URL: a.URL, OptionID: option.ID, Title: a.Title, Platform: a.Platform, Kind: option.Kind, Label: option.Label, Extension: option.Extension, CreatedAt: now, ExpiresAt: now.Add(s.Config.Retention)})
			if err != nil {
				s.jobError(w, err)
				return
			}
			s.Manager.Notify()
			writeJSON(w, 202, j)
			return
		}
	}
	problem(w, 400, "invalid_format", "This format is not offered by the analysis.")
}
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	list, err := s.Manager.Store.List(s.owner(w, r))
	if err != nil {
		s.jobError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"jobs": list})
}
func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	j, err := s.Manager.Store.Get(r.PathValue("id"), s.owner(w, r))
	if err != nil {
		s.jobError(w, err)
		return
	}
	writeJSON(w, 200, j)
}
func (s *Server) cancel(w http.ResponseWriter, r *http.Request) {
	j, err := s.Manager.Cancel(r.PathValue("id"), s.owner(w, r))
	if err != nil {
		s.jobError(w, err)
		return
	}
	writeJSON(w, 200, j)
}
func (s *Server) retry(w http.ResponseWriter, r *http.Request) {
	owner := s.owner(w, r)
	j, err := s.Manager.Store.Get(r.PathValue("id"), owner)
	if err != nil {
		s.jobError(w, err)
		return
	}
	if j.State != "failed" && j.State != "canceled" {
		s.jobError(w, jobs.ErrState)
		return
	}
	if time.Now().After(j.ExpiresAt) {
		problem(w, 410, "job_expired", "This job expired. Analyze the link again.")
		return
	}
	now := time.Now().UTC()
	j.ID = ""
	j.CreatedAt = now
	j.ExpiresAt = now.Add(s.Config.Retention)
	j.Error = ""
	next, err := s.Manager.Store.Create(j)
	if err != nil {
		s.jobError(w, err)
		return
	}
	s.Manager.Notify()
	writeJSON(w, 202, next)
}
func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	err := s.Manager.Delete(r.PathValue("id"), s.owner(w, r))
	if err != nil {
		s.jobError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) jobError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, jobs.ErrNotFound):
		problem(w, 404, "not_found", "Job not found.")
	case errors.Is(err, jobs.ErrCapacity):
		problem(w, 429, "queue_full", err.Error())
	case errors.Is(err, jobs.ErrState):
		problem(w, 409, "invalid_state", err.Error())
	default:
		slog.Error("job operation failed", "error", err)
		problem(w, 500, "operation_failed", "The operation could not be completed. Try again shortly.")
	}
}
func (s *Server) file(w http.ResponseWriter, r *http.Request) {
	j, err := s.Manager.Store.Get(r.PathValue("id"), s.owner(w, r))
	if err != nil {
		s.jobError(w, err)
		return
	}
	if j.State != "complete" || time.Now().After(j.ExpiresAt) {
		problem(w, 410, "file_unavailable", "The file is not ready or has expired.")
		return
	}
	index, err := strconv.Atoi(r.PathValue("file"))
	if err != nil || index < 0 || index >= len(j.Files) {
		s.jobError(w, jobs.ErrNotFound)
		return
	}
	file := j.Files[index]
	if filepath.Base(file.Name) != file.Name {
		s.jobError(w, jobs.ErrNotFound)
		return
	}
	path := filepath.Join(s.Manager.Root, j.ID, file.Name)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		s.jobError(w, jobs.ErrNotFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		s.jobError(w, jobs.ErrNotFound)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	name := safeFilename(j.Title) + filepath.Ext(file.Name)
	if len(j.Files) > 1 {
		name = fmt.Sprintf("%s-%d%s", safeFilename(j.Title), index+1, filepath.Ext(file.Name))
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	http.ServeContent(w, r, name, info.ModTime(), f)
}
func safeFilename(title string) string {
	var b strings.Builder
	for _, r := range title {
		if r == '/' || r == '\\' || r < 32 || strings.ContainsRune(`<>:"|?*`, r) {
			continue
		}
		b.WriteRune(r)
		if b.Len() >= 100 {
			break
		}
	}
	value := strings.Trim(b.String(), " .")
	if value == "" {
		return "download"
	}
	return value
}
func (s *Server) thumbnail(w http.ResponseWriter, r *http.Request) {
	a, ok := s.cached(r.PathValue("id"), s.owner(w, r))
	if !ok || a.Thumbnail == "" {
		http.NotFound(w, r)
		return
	}
	e, ok := s.Manager.Engine.(*media.YTDLP)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, err := security.Parse(a.Thumbnail); err != nil {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", a.Thumbnail, nil)
	resp, err := e.Client.Do(req)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		http.NotFound(w, r)
		return
	}
	// Thumbnails are small, bounded assets. Media downloads use streaming instead.
	data, err := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
	if err != nil || len(data) > 2<<20 {
		http.NotFound(w, r)
		return
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(data)
}
