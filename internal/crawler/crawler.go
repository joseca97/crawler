package crawler

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Fetcher struct {
	client      *http.Client
	concurrency int
	maxDepth    int
}

func (f *Fetcher) Start(ctx context.Context, startURLs []string) <-chan Result {
	jobsChan := make(chan Job, 500)
	resultsChan := make(chan Result, 500)
	outChan := make(chan Result, 500)

	var wg sync.WaitGroup

	for i := 0; i < f.concurrency; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			f.worker(ctx, jobsChan, resultsChan)
		}(i)
	}

	go func() {
		visited := make(map[string]bool)

		for _, url := range startURLs {
			visited[url] = true
			jobsChan <- Job{URL: url, Depth: 1}
		}

		activeJobs := len(startURLs)

		for activeJobs > 0 {
			select {
			case <-ctx.Done():
				goto shutdown
			case res := <-resultsChan:
				activeJobs--

				outChan <- res

				if res.Err != nil {
					continue
				}

				for _, link := range res.FoundLinks {
					if !visited[link] {
						if strings.Contains(link, "example") {
							// if strings.Contains(link, "go.dev") || strings.Contains(link, "golang.org") {
							visited[link] = true

							if f.maxDepth > 1 {
								activeJobs++
								select {
								case jobsChan <- Job{URL: link, Depth: 2}:
								case <-ctx.Done():
									goto shutdown
								default:
									activeJobs--
								}
							}
						}
					}
				}
			}
		}

	shutdown:
		close(jobsChan)
		wg.Wait()
		close(outChan)
	}()

	return outChan
}

func (f *Fetcher) worker(ctx context.Context, jobs <-chan Job, results chan<- Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			results <- f.fetch(ctx, job.URL)
		}
	}
}

func NewFetcher(timeout time.Duration, concurrency int, maxDepth int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		concurrency: concurrency,
		maxDepth:    maxDepth,
	}
}

func (f *Fetcher) fetch(ctx context.Context, url string) Result {
	start := time.Now()
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return Result{URL: url, Err: err}
	}

	req.Header.Set("User-Agent", "GoConcurrentCrawler/1.0")

	resp, err := f.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return Result{URL: url, Duration: duration, Err: err}
	}
	defer resp.Body.Close()

	// Simulating found links
	// var discoveredLinks []string
	// if url == "https://golang.org" {
	// 	discoveredLinks = []string{"https://go.dev", "https://github.com"}
	// }

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return Result{
			URL:        url,
			StatusCode: resp.StatusCode,
			Duration:   duration,
		}
	}

	discoveredLinks := extractLinks(url, resp.Body)

	return Result{
		URL:        url,
		FoundLinks: discoveredLinks,
		StatusCode: resp.StatusCode,
		Duration:   duration,
	}
}
