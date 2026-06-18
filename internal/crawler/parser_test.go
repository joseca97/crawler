package crawler

import (
	"strings"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	baseURL := "https://example.com/blog/"
	htmlContent := `
		<html>
			<body>
				<a href="/home">Relative Root</a>
				<a href="post-1">Relative Page</a>
				<a href="https://other.com">External</a>
				<a href="mailto:test@example.com">Email</a>
				<a href="#section">Fragment</a>
				<a href="https://example.com/blog/post-1#top">External with Fragment</a>
			</body>
		</html>
	`

	links := extractLinks(baseURL, strings.NewReader(htmlContent))

	expected := []string{
		"https://example.com/home",
		"https://example.com/blog/post-1",
		"https://other.com",
	}

	if len(links) != len(expected) {
		t.Fatalf("Expected %d links, got %d: %v", len(expected), len(links), links)
	}

	for i, link := range links {
		if link != expected[i] {
			t.Errorf("Expected link %d to be %s, got %s", i, expected[i], link)
		}
	}
}

func TestResolveAbsoluteURL(t *testing.T) {
	base := "https://example.com/dir/page.html"
	
	tests := []struct {
		target   string
		expected string
	}{
		{"/root", "https://example.com/root"},
		{"relative", "https://example.com/dir/relative"},
		{"../parent", "https://example.com/parent"},
		{"http://absolute.com", "http://absolute.com"},
		{"https://absolute.com/path#frag", "https://absolute.com/path"},
		{"javascript:void(0)", ""},
		{"mailto:a@b.com", ""},
		{"ftp://files.com", ""},
	}

	for _, tt := range tests {
		got := resolveAbsoluteURL(base, tt.target)
		if got != tt.expected {
			t.Errorf("resolveAbsoluteURL(%q, %q) = %q; want %q", base, tt.target, got, tt.expected)
		}
	}
}
