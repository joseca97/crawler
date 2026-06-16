package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type CLIConfig struct {
	Seeds           []string
	FileName        string
	Concurrency     int
	TimeoutSec      int
	MaxDepth        int
	ExcludePatterns []string
}

func ParseFlags() CLIConfig {
	var (
		rawURLs         string
		fileName        string
		concurrency     int
		timeoutSec      int
		maxDepth        int
		excludePatterns string
	)

	flag.StringVar(&rawURLs, "urls", "", "Comma-separated list of seed URLs to crawl")
	flag.StringVar(&fileName, "file", "crawl_output.jsonl", "Output file name")
	flag.IntVar(&concurrency, "workers", 3, "Number of concurrent worker threads")
	flag.IntVar(&timeoutSec, "timeout", 5, "HTTP Request timeout in seconds")
	flag.IntVar(&maxDepth, "depth", 2, "Maximum depth for recursive crawling")
	flag.StringVar(&excludePatterns, "exclude", "", "Comma-separated list of patterns to exclude from crawling")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of Concurrent Web Crawler CLI:\n")
		fmt.Fprintf(os.Stderr, "Example: crawler -urls=\"https://example.com,https://go.dev\" -exclude=\"/search/,/tag/\"\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if rawURLs == "" {
		fmt.Fprintln(os.Stderr, "[ERROR] Missing required flag: -urls")
		flag.Usage()
		os.Exit(1)
	}

	seeds := strings.Split(rawURLs, ",")
	for i, url := range seeds {
		seeds[i] = strings.TrimSpace(url)
	}

	var excludes []string
	if excludePatterns != "" {
		excludes = strings.Split(excludePatterns, ",")
		for i, p := range excludes {
			excludes[i] = strings.TrimSpace(p)
		}
	}

	return CLIConfig{
		Seeds:           seeds,
		FileName:        fileName,
		Concurrency:     concurrency,
		TimeoutSec:      timeoutSec,
		MaxDepth:        maxDepth,
		ExcludePatterns: excludes,
	}
}
