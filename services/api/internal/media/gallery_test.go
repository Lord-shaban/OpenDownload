package media

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
)

func galleryFixture(t *testing.T, platform string) Analysis {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "gallery", platform+"-contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw RawInfo
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	a, err := Normalize(raw, "https://source.example.test/post")
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func originalPNG(t *testing.T, fill color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestGalleryContractFixtures(t *testing.T) {
	for _, platform := range []string{"instagram", "tiktok"} {
		t.Run(platform, func(t *testing.T) {
			a := galleryFixture(t, platform)
			if len(a.Options) != 1 || a.Options[0].Kind != "image" || a.Options[0].Extension != "zip" || len(a.Options[0].AssetURLs) != 2 {
				t.Fatalf("gallery choices: %+v", a.Options)
			}
			public, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(public), ".test") || strings.Contains(string(public), "AssetURLs") {
				t.Fatal("internal asset/source URLs leaked into API metadata")
			}
		})
	}
}

func TestGalleryAccessAndItemBounds(t *testing.T) {
	entry := RawInfo{Ext: "png", URL: "http://images.example.test/green.png", Availability: "public"}
	for _, tc := range []struct {
		name string
		raw  RawInfo
	}{
		{"private collection", RawInfo{Availability: "private", Entries: []RawInfo{entry}}},
		{"private entry", RawInfo{Entries: []RawInfo{{Ext: "png", URL: entry.URL, Availability: "needs_auth"}}}},
		{"DRM entry", RawInfo{Entries: []RawInfo{{Ext: "png", URL: entry.URL, DRM: true}}}},
		{"live entry", RawInfo{Entries: []RawInfo{{Ext: "png", URL: entry.URL, Live: true}}}},
		{"mixed video", RawInfo{Entries: []RawInfo{entry, {Ext: "mp4", URL: entry.URL}}}},
		{"missing asset", RawInfo{Entries: []RawInfo{{Ext: "png"}}}},
		{"too many", RawInfo{Entries: make([]RawInfo, 21)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Normalize(tc.raw, "https://source.example.test/post"); err == nil {
				t.Fatal("unsafe collection was offered")
			}
		})
	}
	entries := make([]RawInfo, 20)
	for i := range entries {
		entries[i] = entry
	}
	if _, err := Normalize(RawInfo{Entries: entries}, "https://source.example.test/post"); err != nil {
		t.Fatal("valid collection at the item bound rejected", err)
	}
}

func TestGalleryZIPPreservesImagesThroughProxy(t *testing.T) {
	green := originalPNG(t, color.RGBA{G: 200, A: 255})
	blue := originalPNG(t, color.RGBA{B: 200, A: 255})
	payloads := map[string][]byte{
		"http://images.example.test/green.png": green,
		"http://cdn.example.test/blue.png":     blue,
	}
	var requests atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, ok := payloads[r.RequestURI]
		if !ok || r.Method != http.MethodGet {
			t.Errorf("unexpected proxy request: %s %s", r.Method, r.RequestURI)
			http.Error(w, "unexpected asset", 404)
			return
		}
		requests.Add(1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer proxy.Close()
	e, err := NewYTDLP("unused", proxy.URL, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Client.CloseIdleConnections()
	a := galleryFixture(t, "instagram")
	dir := t.TempDir()
	var progress []Progress
	outputs, err := e.Download(context.Background(), a, a.Options[0], dir, func(p Progress) { progress = append(progress, p) })
	if err != nil || len(outputs) != 1 || outputs[0].Name != "gallery.zip" || requests.Load() != 2 {
		t.Fatalf("gallery download: %+v %v; requests=%d", outputs, err, requests.Load())
	}
	archive, err := zip.OpenReader(outputs[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	if len(archive.File) != 2 {
		t.Fatalf("archive entries: %d", len(archive.File))
	}
	for i, expected := range [][]byte{green, blue} {
		file := archive.File[i]
		if file.Name != []string{"image-01.png", "image-02.png"}[i] {
			t.Fatalf("unsafe or reordered archive path: %s", file.Name)
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil || !bytes.Equal(data, expected) {
			t.Fatal("original image bytes changed", err)
		}
		if _, err := png.Decode(bytes.NewReader(data)); err != nil {
			t.Fatal("archive member is not a decodable PNG", err)
		}
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 1 || len(progress) != 2 || progress[1].Percent != 99 {
		t.Fatalf("scratch or progress contract: %v %+v %v", files, progress, err)
	}
}

func TestGalleryDownloadFailureBounds(t *testing.T) {
	data := originalPNG(t, color.RGBA{R: 180, A: 255})
	for _, tc := range []struct {
		name   string
		limit  int64
		items  int
		status int
		body   []byte
	}{
		{"aggregate budget", int64(len(data) * 3), 2, 200, data},
		{"item bound before requests", 4096, 21, 200, data},
		{"access denied", 4096, 2, 403, data},
		{"HTML disguised as image", 4096, 2, 200, []byte("<html>access page</html>")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.Header().Set("Content-Type", "image/png")
				w.WriteHeader(tc.status)
				_, _ = w.Write(tc.body)
			}))
			defer proxy.Close()
			e, err := NewYTDLP("unused", proxy.URL, tc.limit)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Client.CloseIdleConnections()
			assets := make([]string, tc.items)
			for i := range assets {
				assets[i] = "http://images.example.test/image.png"
			}
			outputs, err := e.Download(context.Background(), Analysis{}, Option{AssetURLs: assets}, t.TempDir(), func(Progress) {})
			if err == nil || len(outputs) != 0 {
				t.Fatal("failed gallery returned downloadable output", outputs, err)
			}
			if tc.items > 20 && requests.Load() != 0 {
				t.Fatal("item limit checked after network requests")
			}
			if tc.status == 403 && requests.Load() != 1 {
				t.Fatal("continued downloading after access denial")
			}
		})
	}
}

func TestGalleryPrivateAssetDeniedByGuardedProxy(t *testing.T) {
	proxy := httptest.NewServer(security.NewProxy(security.Policy{}))
	defer proxy.Close()
	e, err := NewYTDLP("unused", proxy.URL, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Client.CloseIdleConnections()
	for _, source := range []string{"http://127.0.0.1/image.png", "http://169.254.169.254/image.png"} {
		outputs, err := e.Download(context.Background(), Analysis{}, Option{AssetURLs: []string{source}}, t.TempDir(), func(Progress) {})
		if err == nil || len(outputs) != 0 {
			t.Fatal("gallery accepted a private asset", source, outputs, err)
		}
	}
}

func TestGalleryCancellationStopsBeforeNextImage(t *testing.T) {
	data := originalPNG(t, color.RGBA{G: 180, A: 255})
	started := make(chan struct{}, 1)
	var requests atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
		w.(http.Flusher).Flush()
		started <- struct{}{}
		<-r.Context().Done()
	}))
	defer proxy.Close()
	e, err := NewYTDLP("unused", proxy.URL, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Client.CloseIdleConnections()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		outputs, err := e.Download(ctx, Analysis{}, Option{AssetURLs: []string{"http://images.example.test/one.png", "http://images.example.test/two.png"}}, t.TempDir(), func(Progress) {})
		if len(outputs) != 0 {
			t.Error("canceled gallery returned files")
		}
		done <- err
	}()
	select {
	case <-started:
		cancel()
	case <-ctx.Done():
		t.Fatal("download never reached proxy")
	}
	select {
	case err := <-done:
		if err == nil || requests.Load() != 1 {
			t.Fatal("cancellation did not stop gallery", requests.Load(), err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled gallery did not return promptly")
	}
}
