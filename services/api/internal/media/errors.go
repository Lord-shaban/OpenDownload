package media

import "errors"

// Public categories deliberately omit extractor stderr, URLs and credentials.
var (
	ErrPlatformVerification = errors.New("the source platform requires human verification from this server; the link may still be public")
	ErrUpstreamForbidden    = errors.New("the source platform refused this server's request; the link may still be public")
	ErrUpstreamRateLimit    = errors.New("the source platform temporarily rate-limited this server")
	ErrSourceConnection     = errors.New("this server could not connect to the media source through its egress proxy")
	ErrSourceMetadata       = errors.New("the source platform did not return downloadable media metadata to this server")
)
