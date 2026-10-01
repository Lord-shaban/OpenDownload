package media

import (
	"errors"
	"net/url"
	"strings"
)

// YouTube is deliberately outside the verified 0.1 release scope.
var ErrYouTubeUnavailable = errors.New("YouTube is not supported in OpenDownload 0.1. Try a link from the supported sites")

func YouTubeSource(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	for _, domain := range []string{"youtube.com", "youtu.be", "youtube-nocookie.com"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}
