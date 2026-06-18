package cli

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     CLIConfig
		wantErr  bool
	}{
		{
			name: "basic flags",
			args: []string{"-urls", "https://example.com", "-workers", "5"},
			want: CLIConfig{
				Seeds:       []string{"https://example.com"},
				FileName:    "crawl_output.jsonl",
				Concurrency: 5,
				TimeoutSec:  5,
				MaxDepth:    2,
			},
			wantErr: false,
		},
		{
			name: "multiple urls and excludes",
			args: []string{"-urls", "a.com, b.com", "-exclude", "bad.com, /secret"},
			want: CLIConfig{
				Seeds:           []string{"a.com", "b.com"},
				FileName:        "crawl_output.jsonl",
				Concurrency:     3,
				TimeoutSec:      5,
				MaxDepth:        2,
				ExcludePatterns: []string{"bad.com", "/secret"},
			},
			wantErr: false,
		},
		{
			name: "all flags",
			args: []string{
				"-urls", "test.com",
				"-file", "out.jsonl",
				"-workers", "10",
				"-timeout", "30",
				"-depth", "5",
				"-exclude", "x",
			},
			want: CLIConfig{
				Seeds:           []string{"test.com"},
				FileName:        "out.jsonl",
				Concurrency:     10,
				TimeoutSec:      30,
				MaxDepth:        5,
				ExcludePatterns: []string{"x"},
			},
			wantErr: false,
		},
		{
			name:    "missing urls",
			args:    []string{"-workers", "5"},
			want:    CLIConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags("test", tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
