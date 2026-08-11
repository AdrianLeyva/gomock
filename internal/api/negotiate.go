package api

import (
	"strconv"
	"strings"
)

// prefersHTML reports whether a client explicitly asked for HTML in
// preference to JSON, which in practice means "is this a browser?".
//
// The rule is deliberately asymmetric: wildcard media ranges (*/* and
// application/*) count towards the JSON score but never towards the HTML
// score. That asymmetry is the whole point — it is what keeps
// `curl localhost:8080/`, which sends `Accept: */*`, receiving the JSON
// discovery document, while a browser's
// `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`
// resolves to HTML. A missing Accept header scores zero on both sides and
// therefore also yields JSON.
func prefersHTML(accept string) bool {
	htmlQ := acceptQuality(accept, "text/html", "application/xhtml+xml")
	jsonQ := acceptQuality(accept, "application/json", "application/*", "*/*")

	// A tie goes to HTML: the client named text/html explicitly and rated
	// it no lower than anything JSON-ish, so it wants the page.
	return htmlQ > 0 && htmlQ >= jsonQ
}

// acceptQuality returns the highest q-value among the media ranges in want
// that appear in the Accept header, or 0 if none of them do. Matching is
// exact on the media range string; no wildcard expansion is performed, so
// the caller decides which wildcards (if any) it is willing to honour.
func acceptQuality(accept string, want ...string) float64 {
	best := 0.0

	for _, part := range strings.Split(accept, ",") {
		mediaRange, q := parseAcceptPart(part)
		if mediaRange == "" || q <= 0 {
			continue
		}
		for _, candidate := range want {
			if mediaRange == candidate && q > best {
				best = q
			}
		}
	}

	return best
}

// parseAcceptPart splits a single Accept header element into its
// lower-cased media range and its q-value, defaulting to q=1.0 when the
// parameter is absent or malformed and clamping it to [0, 1].
func parseAcceptPart(part string) (string, float64) {
	segments := strings.Split(part, ";")

	mediaRange := strings.ToLower(strings.TrimSpace(segments[0]))
	if mediaRange == "" {
		return "", 0
	}

	q := 1.0
	for _, param := range segments[1:] {
		param = strings.TrimSpace(param)
		if !strings.HasPrefix(strings.ToLower(param), "q=") {
			continue
		}

		v, err := strconv.ParseFloat(strings.TrimSpace(param[len("q="):]), 64)
		if err != nil {
			break // malformed q: treat the range as unweighted rather than dropping it
		}
		switch {
		case v < 0:
			q = 0
		case v > 1:
			q = 1
		default:
			q = v
		}
		break
	}

	return mediaRange, q
}
