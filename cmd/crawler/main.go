package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"crawler/internal/crawler"
	"crawler/internal/crawler/cli"
)

func main() {
	cfg := cli.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		fmt.Println("\n[SHUTDOWN] Signal received! Cancelling outstanding tasks gracefully...")
		cancel()
	}()

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	fetcher := crawler.NewFetcher(timeout, cfg.Concurrency, cfg.MaxDepth, cfg.ExcludePatterns)

	fmt.Println("Starting Crawler CLI Engine...")
	fmt.Printf("Configurations -> Workers: %d | Timeout: %s | Max Depth: %d | Excludes: %d\n", cfg.Concurrency, timeout, cfg.MaxDepth, len(cfg.ExcludePatterns))
	fmt.Println("--------------------------------------------------")

	store := crawler.NewJSONLStorage(cfg.FileName, 100)
	defer store.Close()

	results := fetcher.Start(ctx, cfg.Seeds)

	for res := range results {
		if res.Err != nil {
			fmt.Printf("[ERROR]   %s: %v\n", res.URL, res.Err)
		} else {
			fmt.Printf("[SUCCESS] %s -> Found %d links (took %v)\n", res.URL, len(res.FoundLinks), res.Duration)
		}

		store.Write(res)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("Crawl pipeline complete. Exiting cleanly")
}
