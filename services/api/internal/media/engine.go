package media

import (
	"archive/zip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
)

type YTDLP struct {
	Binary        string
	Proxy         string
	MaxBytes      int64
	Client        *http.Client
	youtubeSlots  chan struct{}
	diagnosticKey *rsa.PublicKey
}

func NewYTDLP(binary, proxy string, max int64) (*YTDLP, error) {
	u, err := url.Parse(proxy)
	if err != nil || u.Host == "" || u.Scheme != "http" {
		return nil, errors.New("an operator-configured HTTP egress proxy is required")
	}
	client := &http.Client{Timeout: 15 * time.Minute, Transport: &http.Transport{Proxy: http.ProxyURL(u), TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 20 * time.Second}, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		_, err := security.Parse(req.URL.String())
		return err
	}}
	key, err := diagnosticKey()
	if err != nil {
		return nil, err
	}
	return &YTDLP{Binary: binary, Proxy: proxy, MaxBytes: max, Client: client, youtubeSlots: make(chan struct{}, 1), diagnosticKey: key}, nil
}

// A YouTube extraction may start a Node attestation process. Share one slot
// between analysis and downloads so anonymous visitors cannot multiply those
// processes beyond the small cloud worker's memory budget. Waiting is canceled
// with the original request/job context.
func (e *YTDLP) acquireYouTube(ctx context.Context, source string) (func(), error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	host := strings.ToLower(u.Hostname())
	if !youtubeHost(host) || e.youtubeSlots == nil {
		return func() {}, nil
	}
	select {
	case e.youtubeSlots <- struct{}{}:
		return func() { <-e.youtubeSlots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func youtubeHost(host string) bool {
	return host == "youtu.be" || host == "youtube.com" || strings.HasSuffix(host, ".youtube.com") || host == "youtube-nocookie.com" || strings.HasSuffix(host, ".youtube-nocookie.com")
}

func youtubeScratch(cmd *exec.Cmd, source string) (func(), error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	if !youtubeHost(strings.ToLower(u.Hostname())) {
		return func() {}, nil
	}
	dir, err := os.MkdirTemp("", "od-youtube-")
	if err != nil {
		return nil, err
	}
	root, err := filepath.Abs(os.TempDir())
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, errors.New("unsafe extractor scratch directory")
	}
	cmd.Env = append(cmd.Env, "TMPDIR="+dir, "TEMP="+dir, "TMP="+dir)
	// The parent cleans browser profiles even after the child group is killed.
	return func() { _ = os.RemoveAll(dir) }, nil
}

func (e *YTDLP) common() []string {
	// Include public web clients whose HLS/embed formats can differ from the
	// engine defaults. Analysis and download must use the same client policy.
	// These clients do not import accounts or override upstream access checks.
	args := []string{"--ignore-config", "--no-plugin-dirs", "--no-cache-dir", "--no-playlist", "--no-warnings", "--proxy", e.Proxy, "--socket-timeout", "15", "--retries", "2", "--fragment-retries", "2", "--concurrent-fragments", "1", "--hls-prefer-native", "--downloader", "dash:native", "--js-runtimes", "node", "--no-remote-components", "--extractor-args", "youtube:player_client=default,web_safari,web_embedded", "--postprocessor-args", "ffmpeg_i:-protocol_whitelist file,pipe"}
	if e.diagnosticKey != nil {
		args = append(args, "--verbose")
	}
	return args
}

func (e *YTDLP) stderrBuffer() *boundedBuffer {
	limit := 16 << 10
	if e.diagnosticKey != nil {
		limit = 128 << 10 // Verbose diagnostics stay bounded and encrypted.
	}
	return &boundedBuffer{limit: limit}
}

func (e *YTDLP) Analyze(ctx context.Context, source string) (Analysis, error) {
	u, err := security.Parse(source)
	if err != nil {
		return Analysis{}, err
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(u.Path)), ".")
	if imageExt(ext) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodHead, source, nil)
		resp, err := e.Client.Do(req)
		if err != nil {
			return Analysis{}, errors.New("image source could not be reached")
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/") {
			return Analysis{}, ErrUnsupported
		}
		if resp.ContentLength > e.MaxBytes {
			return Analysis{}, errors.New("this image exceeds the instance file limit")
		}
		return Analysis{Title: filepath.Base(u.Path), Platform: u.Hostname(), URL: source, Options: []Option{{ID: "image", Kind: "image", Label: "Original image", Extension: ext, Detail: "Original source file", Bytes: resp.ContentLength, AssetURL: source}}}, nil
	}
	release, err := e.acquireYouTube(ctx, source)
	if err != nil {
		return Analysis{}, err
	}
	defer release()
	args := append(e.common(), "--dump-single-json", "--skip-download", "--", source)
	cmd := command(ctx, e.Binary, args...)
	cleanup, err := youtubeScratch(cmd, source)
	if err != nil {
		return Analysis{}, err
	}
	defer cleanup()
	out := &boundedBuffer{limit: 8 << 20}
	stderr := e.stderrBuffer()
	cmd.Stdout = out
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		e.diagnose("analysis", stderr.Bytes())
		return Analysis{}, extractError(ctx, stderr.Bytes())
	}
	var info RawInfo
	if json.Unmarshal(out.Bytes(), &info) != nil {
		return Analysis{}, errors.New("the extractor returned invalid metadata")
	}
	return Normalize(info, source)
}

func extractError(ctx context.Context, stderr []byte) error {
	if ctx.Err() != nil {
		return errors.New("analysis or processing was canceled or timed out")
	}
	msg := strings.ToLower(string(stderr))
	if strings.Contains(msg, "tunnel connection failed") || strings.Contains(msg, "unable to connect to proxy") {
		return ErrSourceConnection
	}
	if strings.Contains(msg, "not a bot") || strings.Contains(msg, "verify you are human") {
		return ErrPlatformVerification
	}
	if strings.Contains(msg, "http error 429") || strings.Contains(msg, "too many requests") {
		return ErrUpstreamRateLimit
	}
	if strings.Contains(msg, "http error 403") || strings.Contains(msg, "403 forbidden") || strings.Contains(msg, "403: forbidden") {
		return ErrUpstreamForbidden
	}
	if strings.Contains(msg, "sign in") || strings.Contains(msg, "log in") || strings.Contains(msg, "login") || strings.Contains(msg, "logged-in") || strings.Contains(msg, "private video") || strings.Contains(msg, "drm") {
		return ErrAccess
	}
	if strings.Contains(msg, "unsupported url") {
		return ErrUnsupported
	}
	if strings.Contains(msg, "unable to extract") || strings.Contains(msg, "failed to parse json") || strings.Contains(msg, "jsondecodeerror") {
		return ErrSourceMetadata
	}
	return errors.New("the source could not be processed; check the public link or try again later")
}

type progressWriter struct {
	mu      sync.Mutex
	pending string
	update  func(Progress)
}

func (p *progressWriter) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pending += string(data)
	for {
		idx := strings.IndexByte(p.pending, '\n')
		if idx < 0 {
			break
		}
		line := p.pending[:idx]
		p.pending = p.pending[idx+1:]
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "ODPROGRESS" {
			done, _ := strconv.ParseFloat(fields[1], 64)
			total, _ := strconv.ParseFloat(fields[2], 64)
			if total <= 0 {
				total, _ = strconv.ParseFloat(fields[3], 64)
			}
			percent := -1.0
			if total > 0 {
				percent = min(99, done/total*100)
			}
			p.update(Progress{Percent: percent, Phase: "downloading"})
		}
		if strings.HasPrefix(line, "ODFILE ") {
			p.update(Progress{Percent: 99, Phase: "processing"})
		}
	}
	if len(p.pending) > 8192 {
		p.pending = ""
	}
	return len(data), nil
}

func (e *YTDLP) Download(ctx context.Context, a Analysis, opt Option, dir string, update func(Progress)) ([]Output, error) {
	if opt.AssetURL != "" {
		return e.downloadImage(ctx, opt.AssetURL, dir, "media", update)
	}
	if len(opt.AssetURLs) > 0 {
		if len(opt.AssetURLs) > 20 {
			return nil, ErrUnsupported
		}
		file, err := os.Create(filepath.Join(dir, "gallery.zip"))
		if err != nil {
			return nil, err
		}
		z := zip.NewWriter(file)
		var total int64
		for i, source := range opt.AssetURLs {
			outputs, err := e.downloadImage(ctx, source, dir, fmt.Sprintf("image-%02d", i+1), func(Progress) {})
			if err != nil {
				_ = z.Close()
				_ = file.Close()
				return nil, err
			}
			image := outputs[0]
			total += image.Bytes
			if total > e.MaxBytes/2 {
				_ = z.Close()
				_ = file.Close()
				return nil, errors.New("gallery exceeds storage budget")
			}
			f, err := os.Open(image.Path)
			if err != nil {
				_ = z.Close()
				_ = file.Close()
				return nil, err
			}
			entry, err := z.Create(image.Name)
			if err == nil {
				_, err = io.Copy(entry, f)
			}
			_ = f.Close()
			_ = os.Remove(image.Path)
			if err != nil {
				_ = z.Close()
				_ = file.Close()
				return nil, err
			}
			update(Progress{Percent: float64(i+1) / float64(len(opt.AssetURLs)) * 99, Phase: "downloading"})
		}
		err = z.Close()
		closeErr := file.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return collectOutputs(dir, e.MaxBytes)
	}
	release, err := e.acquireYouTube(ctx, a.URL)
	if err != nil {
		return nil, err
	}
	defer release()
	args := append(e.common(), "--newline", "--no-simulate", "--progress", "--progress-template", "download:ODPROGRESS %(progress.downloaded_bytes)s %(progress.total_bytes)s %(progress.total_bytes_estimate)s", "--print", "after_move:ODFILE %(filepath)s", "--max-filesize", strconv.FormatInt(e.MaxBytes, 10), "--restrict-filenames", "--no-overwrites", "--no-mtime", "--paths", dir, "--output", "media.%(ext)s")
	if opt.Kind == "subtitle" {
		args = append(args, "--skip-download", "--sub-langs", opt.Language, "--sub-format", "vtt")
		if opt.Automatic {
			args = append(args, "--write-auto-subs")
		} else {
			args = append(args, "--write-subs")
		}
	} else {
		if !validSelector(opt.Selector) {
			return nil, ErrUnsupported
		}
		args = append(args, "--format", opt.Selector)
		if opt.ID == "audio-mp3" {
			args = append(args, "--extract-audio", "--audio-format", "mp3", "--audio-quality", "192K")
		} else if strings.Contains(opt.Selector, "+") {
			args = append(args, "--merge-output-format", opt.Extension)
		}
	}
	args = append(args, "--", a.URL)
	cmd := command(ctx, e.Binary, args...)
	cleanup, err := youtubeScratch(cmd, a.URL)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	cmd.Stdout = &progressWriter{update: update}
	stderr := e.stderrBuffer()
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		e.diagnose("download", stderr.Bytes())
		return nil, extractError(ctx, stderr.Bytes())
	}
	update(Progress{Percent: 99, Phase: "processing"})
	return collectOutputs(dir, e.MaxBytes)
}

func validSelector(value string) bool {
	parts := strings.Split(value, "+")
	if len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if !safeID.MatchString(p) {
			return false
		}
	}
	return len(parts) > 0
}

func (e *YTDLP) downloadImage(ctx context.Context, source, dir, stem string, update func(Progress)) ([]Output, error) {
	if _, err := security.Parse(source); err != nil {
		return nil, err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, errors.New("image source could not be reached")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, ErrAccess
	}
	reader := io.LimitReader(resp.Body, e.MaxBytes+1)
	header := make([]byte, 512)
	n, err := io.ReadFull(reader, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	header = header[:n]
	ext := ""
	switch http.DetectContentType(header) {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/webp":
		ext = "webp"
	case "image/gif":
		ext = "gif"
	default:
		return nil, errors.New("source is not a supported image")
	}
	path := filepath.Join(dir, stem+"."+ext)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(header)
	if err == nil {
		_, err = io.Copy(f, reader)
	}
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > e.MaxBytes {
		return nil, errors.New("image exceeds file limit")
	}
	update(Progress{Percent: 99, Phase: "processing"})
	return []Output{{Name: filepath.Base(path), Path: path, Bytes: info.Size()}}, nil
}

func collectOutputs(dir string, limit int64) ([]Output, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := []Output{}
	var total int64
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return nil, errors.New("unsafe extractor output")
		}
		name := entry.Name()
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil, errors.New("unsafe output file")
		}
		// Scratch fragments also consume storage even when they are not downloadable.
		total += info.Size()
		if total > limit {
			return nil, errors.New("processed files exceed instance size limit")
		}
		ext := strings.TrimPrefix(filepath.Ext(name), ".")
		switch ext {
		case "mp4", "webm", "mkv", "m4a", "mp3", "ogg", "opus", "flac", "wav", "vtt", "jpg", "jpeg", "png", "gif", "webp", "zip":
		default:
			continue
		}
		out = append(out, Output{Name: name, Path: filepath.Join(dir, name), Bytes: info.Size()})
	}
	if len(out) == 0 {
		return nil, errors.New("the extractor produced no downloadable files")
	}
	return out, nil
}

// Outputs exposes validation to the worker and tests without exposing arbitrary paths.
func Outputs(dir string, limit int64) ([]Output, error) { return collectOutputs(dir, limit) }
