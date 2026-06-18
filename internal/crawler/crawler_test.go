package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShouldExclude(t *testing.T) {
	f := &Fetcher{
		excludePatterns: []string{"/admin", "/private", ".pdf"},
	}

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/admin/settings", true},
		{"https://example.com/blog/post-1", false},
		{"https://example.com/downloads/manual.pdf", true},
		{"https://example.com/private/data", true},
		{"https://example.com/public", false},
	}

	for _, tt := range tests {
		if got := f.shouldExclude(tt.url); got != tt.expected {
			t.Errorf("shouldExclude(%q) = %v; want %v", tt.url, got, tt.expected)
		}
	}
}

func TestHandleResult(t *testing.T) {
	f := &Fetcher{
		maxDepth:    2,
		allowedHost: "example.com",
	}

	visited := make(map[string]bool)
	queue := []Job{}
	activeJobs := 1

	// Test successful result with new links
	res := Result{
		URL:        "http://example.com",
		Depth:      1,
		StatusCode: 200,
		FoundLinks: []string{"http://example.com/a", "http://external.com", "http://example.com/b"},
	}

	f.handleResult(res, visited, &queue, &activeJobs)

	if len(queue) != 2 {
		t.Errorf("Expected 2 jobs in queue, got %d", len(queue))
	}
	if !visited["http://example.com/a"] || !visited["http://example.com/b"] {
		t.Error("Expected internal links to be marked visited")
	}
	if visited["http://external.com"] {
		t.Error("External link should not be marked visited or added to queue")
	}
	if activeJobs != 3 { // 1 (initial) + 2 (new)
		t.Errorf("Expected activeJobs to be 3, got %d", activeJobs)
	}

	// Test depth limit
	resDepth := Result{
		URL:        "http://example.com/a",
		Depth:      2,
		StatusCode: 200,
		FoundLinks: []string{"http://example.com/c"},
	}
	f.handleResult(resDepth, visited, &queue, &activeJobs)
	if len(queue) != 2 { // Should not add new links because nextDepth (3) > maxDepth (2)
		t.Errorf("Expected queue length to stay 2 due to depth limit, got %d", len(queue))
	}
}

func TestCrawlerStart(t *testing.T) {
	// Mock HTTP Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `<html><body><a href="/a">Link A</a><a href="/b">Link B</a></body></html>`)
		case "/a":
			fmt.Fprint(w, `<html><body><a href="/c">Link C</a></body></html>`)
		case "/b":
			fmt.Fprint(w, `<html><body>Done</body></html>`)
		case "/c":
			fmt.Fprint(w, `<html><body>Final</body></html>`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	f := NewFetcher(time.Second, 2, 3, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := f.Start(ctx, []string{ts.URL})

	count := 0
	visited := make(map[string]bool)
	for res := range results {
		if res.Err != nil {
			t.Errorf("Unexpected error for %s: %v", res.URL, res.Err)
		}
		visited[res.URL] = true
		count++
	}

	expectedCount := 4
	if count != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, count)
	}

	if !visited[ts.URL+"/"] && !visited[ts.URL] {
		t.Error("Root URL not visited")
	}
}
