package crawler

import (
	"context"
	"net/http"
	"net/url"
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

	allowedHost := ""
	if len(startURLs) > 0 {
		if u, err := url.Parse(startURLs[0]); err == nil {
			allowedHost = u.Host
		}
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

				nextDepth := res.Depth + 1

				if nextDepth > f.maxDepth {
					continue
				}

				for _, link := range res.FoundLinks {

					parsedLink, err := url.Parse(link)
					if err != nil || parsedLink.Host != allowedHost {
						continue
					}

					if !visited[link] {
						visited[link] = true
						activeJobs++

						select {
						case jobsChan <- Job{URL: link, Depth: nextDepth}:
						case <-ctx.Done():
							goto shutdown
						default:
							activeJobs--
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
			res := f.fetch(ctx, job.URL)
			res.Depth = job.Depth

			results <- res
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
	// reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
