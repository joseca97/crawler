package crawler

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestResultSerialization(t *testing.T) {
	res := Result{
		URL:        "https://example.com",
		FinalURL:   "https://example.com/",
		StatusCode: 200,
		Duration:   123 * time.Millisecond,
		Depth:      1,
		FoundLinks: []string{"a", "b"},
	}

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("Failed to marshal Result: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal Result: %v", err)
	}

	if parsed["duration"] != "123ms" {
		t.Errorf("Expected duration '123ms', got %v", parsed["duration"])
	}
	if parsed["url"] != "https://example.com" {
		t.Errorf("Expected url 'https://example.com', got %v", parsed["url"])
	}
}

func TestJSONLStorage(t *testing.T) {
	tmpFile := "test_output.jsonl"
	defer os.Remove(tmpFile)

	storage := NewJSONLStorage(tmpFile, 2) // Buffer size 2
	
	res1 := Result{URL: "u1", StatusCode: 200}
	res2 := Result{URL: "u2", StatusCode: 404}
	
	storage.Write(res1)
	storage.Write(res2)
	
	// Wait a bit for the async worker to process or just close it
	storage.Close()

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	lines := 0
	for _, line := range os.ExpandEnv(string(content)) {
		if line == '\n' {
			lines++
		}
	}
	// Better way to count lines
	lineCount := 0
	for _, char := range string(content) {
		if char == '\n' {
			lineCount++
		}
	}

	if lineCount != 2 {
		t.Errorf("Expected 2 lines in output file, got %d", lineCount)
	}
}
