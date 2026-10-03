package media

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func socialEngine(t *testing.T, body string, status int) *YTDLP {
	t.Helper()
	e, err := NewYTDLP("unused", "http://127.0.0.1:8090", 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	e.Client.Transport = assetTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}, nil
	})
	return e
}

func TestSocialProgressiveFormatsWithoutCodecs(t *testing.T) {
	for _, platform := range []string{"LinkedIn", "Pinterest", "Threads"} {
		t.Run(platform, func(t *testing.T) {
			raw := RawInfo{Extractor: platform, Formats: []RawFormat{{ID: "0", Ext: "mp4", Protocol: "https", Height: 720}}}
			a, err := Normalize(raw, "https://example.com/post")
			if err != nil || len(a.Options) != 2 || a.Options[0].Kind != "video" || a.Options[0].Label != "720p" || a.Options[0].Detail != "Source file · codec not reported" || a.Options[1].Detail != "Conversion · requires an audio track" {
				t.Fatalf("lost or invented metadata: %+v %v", a, err)
			}
			for _, change := range []func(*RawFormat){
				func(f *RawFormat) { f.DRM = true },
				func(f *RawFormat) { f.Protocol = "file" },
				func(f *RawFormat) { f.ID = "0;bad" },
				func(f *RawFormat) { f.VCodec = "none"; f.ACodec = "none" },
			} {
				candidate := raw
				candidate.Formats = append([]RawFormat(nil), raw.Formats...)
				change(&candidate.Formats[0])
				if _, err := Normalize(candidate, "https://example.com/post"); !errors.Is(err, ErrUnsupported) {
					t.Fatalf("unsafe format accepted: %v", err)
				}
			}
		})
	}
}

func TestThreadsSelectsRequestedPostAndKeepsURLsPrivate(t *testing.T) {
	body := `<script type="application/json">{"data":{"edges":[{"node":{"code":"other","video_versions":[{"url":"https://cdn.example/wrong.mp4"}]}},{"node":{"thread_items":[{"post":{"code":"wanted","user":{"username":"creator"},"caption":{"text":"Public video"},"video_versions":[{"url":"https://cdn.example/right.mp4","height":720,"width":1280}]}}]}}]}}</script>`
	e := socialEngine(t, body, 200)
	a, err := e.Analyze(context.Background(), "https://www.threads.com/@creator/post/wanted")
	if err != nil || a.Platform != "Threads" || a.Creator != "creator" || len(a.Options) != 2 || a.Options[0].MediaURL != "https://cdn.example/right.mp4" || a.Options[1].MediaURL != a.Options[0].MediaURL || a.Options[0].Selector != "best" {
		t.Fatalf("wrong post or selector: %+v %v", a, err)
	}
	public, _ := json.Marshal(a)
	if strings.Contains(string(public), "cdn.example") || strings.Contains(string(public), "MediaURL") {
		t.Fatal("signed URL exposed in public response")
	}
	if _, err := e.Analyze(context.Background(), "https://threads.net/@creator/post/missing"); !errors.Is(err, ErrSourceMetadata) {
		t.Fatal("recommended post substituted for missing post", err)
	}
}

func TestThreadsImagesAndAccessBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, post string
		kind       string
		wantErr    error
	}{
		{"image", `{"code":"wanted","image_versions2":{"candidates":[{"url":"https://cdn.example/small.jpg","width":100,"height":100},{"url":"https://cdn.example/original.jpg","width":1080,"height":1080}]}}`, "image", nil},
		{"gallery", `{"code":"wanted","carousel_media":[{"image_versions2":{"candidates":[{"url":"https://cdn.example/1.jpg"}]}},{"image_versions2":{"candidates":[{"url":"https://cdn.example/2.jpg"}]}}]}`, "gallery", nil},
		{"private", `{"code":"wanted","user":{"is_private":true},"video_versions":[{"url":"https://cdn.example/v.mp4"}]}`, "", ErrAccess},
		{"text", `{"code":"wanted","caption":{"text":"no media"}}`, "", ErrUnsupported},
		{"mixed", `{"code":"wanted","carousel_media":[{"video_versions":[{"url":"https://cdn.example/v.mp4"}]},{"image_versions2":{"candidates":[{"url":"https://cdn.example/1.jpg"}]}}]}`, "", ErrUnsupported},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := socialEngine(t, `<script type="application/json">`+test.post+`</script>`, 200)
			a, err := e.Analyze(context.Background(), "https://threads.net/@creator/post/wanted")
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("wanted %v, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil || len(a.Options) != 1 || a.Options[0].ID != test.kind {
				t.Fatalf("unexpected image result: %+v %v", a, err)
			}
			if test.name == "image" && !strings.Contains(a.Options[0].AssetURL, "original") {
				t.Fatal("smaller image selected")
			}
		})
	}
}

func TestSocialMetadataHTTPBoundaries(t *testing.T) {
	for _, item := range []struct {
		status int
		err    error
	}{{401, ErrAccess}, {403, ErrUpstreamForbidden}, {429, ErrUpstreamRateLimit}, {404, ErrSourceMetadata}} {
		e := socialEngine(t, "", item.status)
		if _, err := e.Analyze(context.Background(), "https://threads.com/@creator/post/wanted"); !errors.Is(err, item.err) {
			t.Fatalf("HTTP %d: %v", item.status, err)
		}
	}
	e := socialEngine(t, strings.Repeat("x", socialMetadataLimit+1), 200)
	if _, err := e.Analyze(context.Background(), "https://threads.com/@creator/post/wanted"); !errors.Is(err, ErrSourceMetadata) {
		t.Fatal("unbounded metadata accepted", err)
	}
	for _, source := range []string{"https://threads.com/@creator", "https://threads.net/search?q=video"} {
		if _, err := e.Analyze(context.Background(), source); !errors.Is(err, ErrUnsupported) {
			t.Fatal("non-post URL accepted", err)
		}
	}
}

func TestPinterestOriginalImageAndVideoDelegation(t *testing.T) {
	body := `{"resource_response":{"data":{"id":"123","title":"A pin","images":{"orig":{"url":"https://i.pinimg.com/originals/photo.jpg"}},"videos":null}}}`
	e := socialEngine(t, body, 200)
	a, err := e.Analyze(context.Background(), "https://www.pinterest.co.uk/pin/123/")
	if err != nil || a.Platform != "Pinterest" || len(a.Options) != 1 || a.Options[0].AssetURL != "https://i.pinimg.com/originals/photo.jpg" {
		t.Fatalf("pin image lost: %+v %v", a, err)
	}
	for _, body := range []string{
		`{"resource_response":{"data":{"id":"123","videos":{"video_list":{"720p":{"url":"https://v.pinimg.com/v.mp4"}}}}}}`,
		`{"resource_response":{"data":{"id":"different","images":{"orig":{"url":"https://i.pinimg.com/photo.jpg"}}}}}`,
	} {
		e = socialEngine(t, body, 200)
		u, _ := url.Parse("https://pinterest.com/pin/123/")
		if _, handled, err := e.analyzePinterestImage(context.Background(), u); handled || err != nil {
			t.Fatal("video or mismatched pin did not delegate to upstream")
		}
	}
}

func TestSocialDomainBoundaries(t *testing.T) {
	for _, host := range []string{"threads.com.attacker.net", "notthreads.net", "pinterest.com.attacker.net", "notpinterest.co.uk"} {
		if threadsHost(host) || pinterestHost(host) {
			t.Fatal("spoofed social domain", host)
		}
	}
}

func TestSocialShortLinksOnlyResolveToTheirOwnPlatform(t *testing.T) {
	for _, item := range []struct {
		source, target string
		ok             bool
	}{
		{"https://pin.it/short", "https://pinterest.co.uk/pin/123/", true},
		{"https://lnkd.in/short", "https://www.linkedin.com/feed/update/urn:li:activity:123", true},
		{"https://threads.net/share/short", "https://threads.com/@creator/post/wanted", true},
		{"https://pin.it/short", "https://pinterest.com.attacker.net/pin/123/", false},
		{"https://lnkd.in/short", "https://example.com/video.mp4", false},
		{"https://threads.com/share/short", "https://threads.com/login", false},
	} {
		e := socialEngine(t, "", 200)
		e.Client.Transport = assetTransport(func(r *http.Request) (*http.Response, error) {
			status := 200
			header := make(http.Header)
			if r.URL.String() == item.source {
				status = 302
				header.Set("Location", item.target)
			}
			return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		})
		u, _ := url.Parse(item.source)
		resolved, err := e.resolveSocialURL(context.Background(), u)
		if item.ok && (err != nil || resolved.String() != item.target) || !item.ok && !errors.Is(err, ErrUnsupported) {
			t.Fatalf("%s => %+v %v", item.source, resolved, err)
		}
	}
}

func TestLinkedInPrefersHighestReportedBitrateWithoutInventingResolution(t *testing.T) {
	a, err := Normalize(RawInfo{Extractor: "LinkedIn", Formats: []RawFormat{
		{ID: "0", Ext: "mp4", Protocol: "https", TBR: 58},
		{ID: "2", Ext: "mp4", Protocol: "https", TBR: 143},
	}}, "https://linkedin.com/posts/sample")
	if err != nil || a.Options[0].Selector != "2" || a.Options[0].Label != "Original video" {
		t.Fatalf("incorrect unprobed selection: %+v %v", a, err)
	}
}
