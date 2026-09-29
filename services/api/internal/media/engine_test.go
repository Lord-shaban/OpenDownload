package media

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type assetTransport func(*http.Request) (*http.Response, error)

func (f assetTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestImagesRejectHTMLAndEnforceStreamLimit(t *testing.T) {
	png, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aD1sAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		name, data string
		limit      int64
		valid      bool
	}{{"image", string(png), 1024, true}, {"html", "<html>not an image</html>", 1024, false}, {"stream-limit", string(png) + strings.Repeat("x", 2000), 1024, false}} {
		t.Run(item.name, func(t *testing.T) {
			e, err := NewYTDLP("unused", "http://127.0.0.1:8090", item.limit)
			if err != nil {
				t.Fatal(err)
			}
			e.Client.Transport = assetTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(item.data)), ContentLength: -1, Header: make(http.Header)}, nil
			})
			outputs, err := e.downloadImage(context.Background(), "https://assets.example/image.png", t.TempDir(), "media", func(Progress) {})
			if item.valid {
				if err != nil || len(outputs) != 1 || outputs[0].Bytes != int64(len(png)) {
					t.Fatalf("invalid image output: %+v %v", outputs, err)
				}
			} else if err == nil {
				t.Fatal("unsafe or oversized image was accepted")
			}
		})
	}
}
func TestOutputBudgetIncludesAbandonedFragments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "media.mp4"), []byte("small"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "media.mp4.part"), make([]byte, 2048), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Outputs(dir, 1024); err == nil {
		t.Fatal("abandoned fragment bypassed storage budget")
	}
}
