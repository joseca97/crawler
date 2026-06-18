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
	cfg, err := parseFlags(os.Args[0], os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func parseFlags(programName string, args []string) (CLIConfig, error) {
	var (
		rawURLs         string
		fileName        string
		concurrency     int
		timeoutSec      int
		maxDepth        int
		excludePatterns string
	)

	fs := flag.NewFlagSet(programName, flag.ContinueOnError)

	fs.StringVar(&rawURLs, "urls", "", "Comma-separated list of seed URLs to crawl")
	fs.StringVar(&fileName, "file", "crawl_output.jsonl", "Output file name")
	fs.IntVar(&concurrency, "workers", 3, "Number of concurrent worker threads")
	fs.IntVar(&timeoutSec, "timeout", 5, "HTTP Request timeout in seconds")
	fs.IntVar(&maxDepth, "depth", 2, "Maximum depth for recursive crawling")
	fs.StringVar(&excludePatterns, "exclude", "", "Comma-separated list of patterns to exclude from crawling")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of Concurrent Web Crawler CLI:\n")
		fmt.Fprintf(os.Stderr, "Example: crawler -urls=\"https://example.com,https://go.dev\" -exclude=\"/search/,/tag/\"\n\n")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return CLIConfig{}, err
	}

	if rawURLs == "" {
		return CLIConfig{}, fmt.Errorf("missing required flag: -urls")
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
	}, nil
}
