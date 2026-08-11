package api

import "testing"

// TestPrefersHTML pins the exact behaviour that decides whether GET /
// returns the console or the discovery document. The rows that matter most
// are the wildcard ones: if `*/*` ever starts resolving to HTML, every
// curl and fetch client silently breaks.
func TestPrefersHTML(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{"missing header", "", false},
		{"curl default wildcard", "*/*", false},
		{"explicit json", "application/json", false},
		{"json with charset param", "application/json; charset=utf-8", false},
		{"application wildcard", "application/*", false},
		{"text wildcard is not html", "text/*", false},
		{"browser", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", true},
		{"bare html", "text/html", true},
		{"xhtml only", "application/xhtml+xml", true},
		{"html refused via q=0", "text/html;q=0", false},
		{"html outranked by json", "text/html;q=0.5,application/json", false},
		{"html outranks wildcard", "text/html;q=0.6,*/*;q=0.2", true},
		{"tie goes to html", "text/html,application/json", true},
		{"spacing and case are ignored", "  TEXT/HTML ;  Q=1.0 , */* ; q=0.8 ", true},
		{"malformed q is treated as 1.0", "text/html;q=banana,*/*;q=0.8", true},
		{"unrelated types only", "image/png,image/webp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prefersHTML(tt.accept); got != tt.want {
				t.Errorf("prefersHTML(%q) = %v, want %v", tt.accept, got, tt.want)
			}
		})
	}
}

func TestAcceptQuality(t *testing.T) {
	tests := []struct {
		accept string
		want   []string
		q      float64
	}{
		{"text/html", []string{"text/html"}, 1.0},
		{"text/html;q=0.4", []string{"text/html"}, 0.4},
		{"text/html;q=0.4,text/html;q=0.9", []string{"text/html"}, 0.9},
		{"application/json", []string{"text/html"}, 0},
		{"text/html;q=5", []string{"text/html"}, 1.0},
		{"text/html;q=-2", []string{"text/html"}, 0},
		{"a/b,application/xhtml+xml;q=0.3", []string{"text/html", "application/xhtml+xml"}, 0.3},
	}

	for _, tt := range tests {
		if got := acceptQuality(tt.accept, tt.want...); got != tt.q {
			t.Errorf("acceptQuality(%q, %v) = %v, want %v", tt.accept, tt.want, got, tt.q)
		}
	}
}
