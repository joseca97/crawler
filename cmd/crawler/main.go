package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	// Replace "mycrawler" with your actual module name from go.mod
	"crawler/internal/crawler"
)

func main() {
	seeds := []string{
		// "https://golang.org",
		// "https://go.dev/doc/",
		"https://example.com",
		// "https://google.com",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		fmt.Println("\n[SHUTDOWN] Signal received! Cancelling outstanding tasks gracefully...")
		cancel()
	}()

	// Initialize with a 5s timeout and a strict pool of 3 workers
	fetcher := crawler.NewFetcher(3*time.Second, 3, 2)

	fmt.Println("Starting Recursive Crawler with State Management...")
	fmt.Println("--------------------------------------------------")

	// Start the engine
	results := fetcher.Start(ctx, seeds)

	// 2. Consume the results as they are completed by the workers
	for res := range results {
		if res.Err != nil {
			fmt.Printf("[ERROR]   %s: %v\n", res.URL, res.Err)
		} else {
			fmt.Printf("[SUCCESS] %s -> Found %d links (took %v)\n", res.URL, len(res.FoundLinks), res.Duration)
		}
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("Stage 3 Complete! All workers exited cleanly.")
}

// func main_two() {
// 	urls := []string{
// 		"https://golang.org",
// 		"https://google.com",
// 		"https://github.com",
// 		"https://thisurlshouldfail12345.com",
// 	}
//
// 	// Initialize the modular fetcher with a 5-second timeout
// 	fetcher := crawler.NewFetcher(5*time.Second, 2)
//
// 	fmt.Println("Starting modular concurrent fetcher...")
// 	fmt.Println("--------------------------------------------------")
//
// 	// Start fetching. We immediately get back a read-only channel.
// 	results := fetcher.FetchAll(urls)
//
// 	// Consume results as they stream in
// 	for res := range results {
// 		if res.Err != nil {
// 			fmt.Printf("[ERROR]   %s: %v (took %v)\n", res.URL, res.Err, res.Duration)
// 		} else {
// 			fmt.Printf("[SUCCESS] %s -> Status: %d (took %v)\n", res.URL, res.StatusCode, res.Duration)
// 		}
// 	}
//
// 	fmt.Println("--------------------------------------------------")
// 	fmt.Println("Stage 1 (Modular) complete!")
// }
