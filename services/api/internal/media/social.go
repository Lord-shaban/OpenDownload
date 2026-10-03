package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Lord-shaban/OpenDownload/services/api/internal/security"
	"golang.org/x/net/html"
)

const socialMetadataLimit = 8 << 20

var threadsPath = regexp.MustCompile(`^/(?:@[^/]+/post|t)/([A-Za-z0-9_-]+)/?$`)
var pinPath = regexp.MustCompile(`^/pin/(?:[\w-]+--)?(\d+)/?$`)

func domainHost(host, domain string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	return host == domain || strings.HasSuffix(host, "."+domain)
}

func threadsHost(host string) bool {
	return domainHost(host, "threads.com") || domainHost(host, "threads.net")
}

func pinterestHost(host string) bool {
	for _, suffix := range strings.Fields("com fr de ch jp cl ca it co.uk nz ru com.au at pt co.kr es com.mx dk ph th com.uy co nl info kr ie vn com.vn ec mx in pe co.at hu co.in co.nz id com.ec com.py tw be uk com.bo com.pe") {
		if domainHost(host, "pinterest."+suffix) {
			return true
		}
	}
	return false
}

// All metadata and redirects use the same guarded transport as media files.
func (e *YTDLP) socialGET(ctx context.Context, source string) ([]byte, *url.URL, error) {
	if _, err := security.Parse(source); err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "OpenDownload/0.1 (+https://github.com/Lord-shaban/OpenDownload)")
	if threadsHost(req.URL.Hostname()) {
		// Threads exposes public post JSON in its search-preview representation.
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	}
	if domainHost(req.URL.Hostname(), "pinterest.com") {
		req.Header.Set("X-Pinterest-PWS-Handler", "www/[username].js")
	}
	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, nil, ErrSourceConnection
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, nil, ErrAccess
	case http.StatusForbidden:
		return nil, nil, ErrUpstreamForbidden
	case http.StatusTooManyRequests:
		return nil, nil, ErrUpstreamRateLimit
	default:
		return nil, nil, ErrSourceMetadata
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, socialMetadataLimit+1))
	if err != nil || len(data) > socialMetadataLimit {
		return nil, nil, ErrSourceMetadata
	}
	return data, resp.Request.URL, nil
}

func (e *YTDLP) resolveSocialURL(ctx context.Context, u *url.URL) (*url.URL, error) {
	host := strings.ToLower(u.Hostname())
	if host != "pin.it" && host != "lnkd.in" && !(threadsHost(host) && strings.HasPrefix(u.Path, "/share/")) {
		return u, nil
	}
	_, resolved, err := e.socialGET(ctx, u.String())
	if err != nil {
		return nil, err
	}
	if host == "pin.it" && pinterestHost(resolved.Hostname()) || host == "lnkd.in" && domainHost(resolved.Hostname(), "linkedin.com") || threadsHost(host) && threadsHost(resolved.Hostname()) && threadsPath.MatchString(resolved.Path) {
		return security.Parse(resolved.String())
	}
	return nil, ErrUnsupported
}

func (e *YTDLP) analyzePinterestImage(ctx context.Context, u *url.URL) (Analysis, bool, error) {
	match := pinPath.FindStringSubmatch(u.Path)
	if match == nil {
		return Analysis{}, true, ErrUnsupported
	}
	options, _ := json.Marshal(map[string]any{"options": map[string]string{"id": match[1], "field_set_key": "unauth_react_main_pin"}})
	endpoint := "https://www.pinterest.com/resource/PinResource/get/?" + url.Values{"data": {string(options)}}.Encode()
	data, _, err := e.socialGET(ctx, endpoint)
	if err != nil {
		// The upstream video extractor may work when this image endpoint is unavailable.
		return Analysis{}, false, nil
	}
	var response struct {
		Response struct {
			Data struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Images      map[string]struct {
					URL string `json:"url"`
				} `json:"images"`
				Videos json.RawMessage `json:"videos"`
				Story  json.RawMessage `json:"story_pin_data"`
			} `json:"data"`
		} `json:"resource_response"`
	}
	if json.Unmarshal(data, &response) != nil || response.Response.Data.ID != match[1] {
		return Analysis{}, false, nil
	}
	pin := response.Response.Data
	for _, field := range []json.RawMessage{pin.Videos, pin.Story} {
		if len(field) > 0 && string(field) != "null" && string(field) != "{}" {
			return Analysis{}, false, nil
		}
	}
	image := pin.Images["orig"].URL
	imageURL, err := security.Parse(image)
	if err != nil || !domainHost(imageURL.Hostname(), "pinimg.com") {
		return Analysis{}, true, ErrUnsupported
	}
	title := pin.Title
	if title == "" {
		title = pin.Description
	}
	a, err := Normalize(RawInfo{Title: title, Extractor: "Pinterest", URL: image, Ext: "jpg"}, u.String())
	return a, true, err
}

type socialImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type threadsPost struct {
	Code string `json:"code"`
	User struct {
		Username string `json:"username"`
		Private  bool   `json:"is_private"`
	} `json:"user"`
	Caption struct {
		Text string `json:"text"`
	} `json:"caption"`
	Videos []socialImage `json:"video_versions"`
	Images struct {
		Candidates []socialImage `json:"candidates"`
	} `json:"image_versions2"`
	Carousel []threadsPost `json:"carousel_media"`
}

func findThreadsPost(value any, code string, depth int) *threadsPost {
	if depth > 80 {
		return nil
	}
	switch node := value.(type) {
	case map[string]any:
		if node["code"] == code {
			data, _ := json.Marshal(node)
			var post threadsPost
			if json.Unmarshal(data, &post) == nil {
				return &post
			}
		}
		for _, child := range node {
			if post := findThreadsPost(child, code, depth+1); post != nil {
				return post
			}
		}
	case []any:
		for _, child := range node {
			if post := findThreadsPost(child, code, depth+1); post != nil {
				return post
			}
		}
	}
	return nil
}

func parseThreadsPage(data []byte, code string) (*threadsPost, error) {
	z := html.NewTokenizer(strings.NewReader(string(data)))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return nil, ErrSourceMetadata
		case html.StartTagToken:
			token := z.Token()
			if token.Data != "script" {
				continue
			}
			isJSON := false
			for _, attr := range token.Attr {
				if attr.Key == "type" && attr.Val == "application/json" {
					isJSON = true
				}
			}
			if !isJSON || z.Next() != html.TextToken {
				continue
			}
			var value any
			if json.Unmarshal(z.Text(), &value) != nil {
				continue
			}
			if post := findThreadsPost(value, code, 0); post != nil {
				return post, nil
			}
		}
	}
}

func bestSocialImage(images []socialImage) string {
	best := socialImage{}
	for _, image := range images {
		if _, err := security.Parse(image.URL); err == nil && (best.URL == "" || int64(image.Width)*int64(image.Height) > int64(best.Width)*int64(best.Height)) {
			best = image
		}
	}
	return best.URL
}

func (e *YTDLP) analyzeThreads(ctx context.Context, u *url.URL) (Analysis, error) {
	match := threadsPath.FindStringSubmatch(u.Path)
	if match == nil {
		return Analysis{}, ErrUnsupported
	}
	data, finalURL, err := e.socialGET(ctx, u.String())
	if err != nil {
		return Analysis{}, err
	}
	finalMatch := threadsPath.FindStringSubmatch(finalURL.Path)
	if !threadsHost(finalURL.Hostname()) || finalMatch == nil || finalMatch[1] != match[1] {
		return Analysis{}, ErrAccess
	}
	post, err := parseThreadsPage(data, match[1])
	if err != nil {
		return Analysis{}, err
	}
	if post.User.Private {
		return Analysis{}, ErrAccess
	}
	items := post.Carousel
	if len(items) == 0 {
		items = []threadsPost{*post}
	}
	if len(items) > 20 {
		return Analysis{}, ErrUnsupported
	}
	raw := RawInfo{Title: post.Caption.Text, Uploader: post.User.Username, Extractor: "Threads"}
	if raw.Title == "" {
		raw.Title = "Threads post " + post.Code
	}
	if len(items) == 1 && len(items[0].Videos) > 0 {
		urls := map[string]string{}
		for _, video := range items[0].Videos {
			if _, err := security.Parse(video.URL); err != nil {
				continue
			}
			id := fmt.Sprintf("threads-%dx%d", video.Width, video.Height)
			urls[id] = video.URL
			raw.Formats = append(raw.Formats, RawFormat{ID: id, Ext: "mp4", Height: video.Height, Width: video.Width, Protocol: "https", URL: video.URL})
		}
		if len(raw.Formats) == 0 {
			return Analysis{}, ErrUnsupported
		}
		raw.Thumbnail = bestSocialImage(items[0].Images.Candidates)
		a, err := Normalize(raw, u.String())
		if err != nil {
			return Analysis{}, err
		}
		for i := range a.Options {
			if mediaURL := urls[a.Options[i].Selector]; mediaURL != "" {
				a.Options[i].MediaURL = mediaURL
				a.Options[i].Selector = "best"
			}
		}
		return a, nil
	}
	for _, item := range items {
		if len(item.Videos) > 0 {
			return Analysis{}, ErrUnsupported
		}
		image := bestSocialImage(item.Images.Candidates)
		if image == "" {
			return Analysis{}, ErrUnsupported
		}
		raw.Entries = append(raw.Entries, RawInfo{URL: image, Ext: "jpg"})
	}
	if len(raw.Entries) == 1 {
		raw.URL = raw.Entries[0].URL
		raw.Ext = "jpg"
		raw.Entries = nil
	}
	return Normalize(raw, u.String())
}
