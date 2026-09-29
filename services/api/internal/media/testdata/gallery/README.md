# Gallery contract fixtures

These JSON documents are original OpenDownload test inputs, licensed under the
repository's MIT license. They model the internal `RawInfo` contract with
Instagram/TikTok platform labels. They are **not captured platform responses**,
gallery-dl output, published posts or proof of anonymous extraction.

All URLs use reserved `.test` names and are served only by a local test proxy.
`gallery_test.go` generates its own green/blue PNGs using Go's image encoder;
no external images, signed CDN URLs, cookies or third-party media are committed.

Tests verify normalized choices, concealed asset URLs, archive content, proxy
routing, item/aggregate byte bounds, access failure and cancellation. See
`docs/GALLERY_EVALUATION.md` for the separate, unrun live-platform gates.
