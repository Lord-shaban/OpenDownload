package media

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestExtractionFailureCategories(t *testing.T) {
	for _, item := range []struct {
		name, diagnostic string
		want             error
	}{
		{"bot-check", "ERROR: [youtube] public-id: Sign in to confirm you're not a bot. Use --cookies", ErrPlatformVerification},
		{"human-check", "Please verify you are human", ErrPlatformVerification},
		{"proxy-403", "Unable to download webpage: Tunnel connection failed: 403 Forbidden", ErrSourceConnection},
		{"proxy-down", "Unable to connect to proxy", ErrSourceConnection},
		{"upstream-403", "Unable to download video data: HTTP Error 403: Forbidden", ErrUpstreamForbidden},
		{"rate-limit", "HTTP Error 429: Too Many Requests", ErrUpstreamRateLimit},
		{"private", "This is a private video. Sign in if you've been granted access", ErrAccess},
		{"drm", "This video is DRM-protected", ErrAccess},
		{"unsupported", "Unsupported URL: https://example.com", ErrUnsupported},
	} {
		t.Run(item.name, func(t *testing.T) {
			err := extractError(context.Background(), []byte(item.diagnostic+" https://cdn.example/video?token=secret"))
			if !errors.Is(err, item.want) || strings.Contains(err.Error(), "token=") {
				t.Fatalf("unsafe or misleading category: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := extractError(ctx, []byte("not a bot")); errors.Is(err, ErrPlatformVerification) {
		t.Fatal("cancellation must take precedence over stale extractor output")
	}
}
