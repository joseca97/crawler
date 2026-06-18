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
	client          *http.Client
	concurrency     int
	maxDepth        int
	allowedHost     string
	excludePatterns []string
}

func (f *Fetcher) shouldExclude(urlStr string) bool {
	for _, p := range f.excludePatterns {
		if strings.Contains(urlStr, p) {
			return true
		}
	}
	return false
}

func (f *Fetcher) Start(ctx context.Context, startURLs []string) <-chan Result {
	jobsChan := make(chan Job, 500)
	resultsChan := make(chan Result, 500)
	outChan := make(chan Result, 500)

	var wg sync.WaitGroup
	f.startWorkers(ctx, &wg, jobsChan, resultsChan)

	if len(startURLs) > 0 {
		if u, err := url.Parse(startURLs[0]); err == nil {
			f.allowedHost = u.Host
		}
	}

	go f.runCoordinator(ctx, startURLs, jobsChan, resultsChan, outChan, &wg)

	return outChan
}

func (f *Fetcher) runCoordinator(ctx context.Context, startURLs []string, jobsChan chan Job, resultsChan chan Result, outChan chan Result, wg *sync.WaitGroup) {
	defer func() {
		close(jobsChan)
		wg.Wait()
		close(outChan)
	}()

	visited := make(map[string]bool)
	var queue []Job

	for _, url := range startURLs {
		visited[url] = true
		queue = append(queue, Job{URL: url, Depth: 1})
	}

	activeJobs := len(startURLs)

	for activeJobs > 0 {
		var sendChan chan<- Job
		var nextJob Job

		if len(queue) > 0 {
			sendChan = jobsChan
			nextJob = queue[0]
		}

		select {
		case <-ctx.Done():
			return

		case sendChan <- nextJob:
			queue = queue[1:]

		case res := <-resultsChan:
			activeJobs--
			outChan <- res
			f.handleResult(res, visited, &queue, &activeJobs)
		}
	}
}

func (f *Fetcher) handleResult(res Result, visited map[string]bool, queue *[]Job, activeJobs *int) {
	if res.FinalURL != "" && res.FinalURL != res.URL {
		visited[res.FinalURL] = true
	}

	if res.Err != nil || res.StatusCode != 200 {
		return
	}

	nextDepth := res.Depth + 1
	if nextDepth > f.maxDepth {
		return
	}

	for _, link := range res.FoundLinks {
		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		if parsedLink.Host != "" && parsedLink.Host != f.allowedHost {
			continue
		}

		if !visited[link] && !f.shouldExclude(link) {
			visited[link] = true
			*activeJobs++
			*queue = append(*queue, Job{URL: link, Depth: nextDepth})
		}
	}
}

func (f *Fetcher) startWorkers(ctx context.Context, wg *sync.WaitGroup, jobsChan chan Job, resultsChan chan Result) {
	for i := 0; i < f.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.worker(ctx, jobsChan, resultsChan)
		}()
	}
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

func NewFetcher(timeout time.Duration, concurrency int, maxDepth int, excludePatterns []string) *Fetcher {
	f := &Fetcher{
		concurrency:     concurrency,
		maxDepth:        maxDepth,
		excludePatterns: excludePatterns,
	}

	f.client = &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			// If allowedHost is set, ensure redirects stay within it
			if f.allowedHost != "" && req.URL.Host != f.allowedHost {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return f
}

func (f *Fetcher) fetch(ctx context.Context, url string) Result {
	start := time.Now()

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

	finalURL := resp.Request.URL.String()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return Result{
			URL:        url,
			FinalURL:   finalURL,
			StatusCode: resp.StatusCode,
			Duration:   duration,
		}
	}

	discoveredLinks := extractLinks(finalURL, resp.Body)

	return Result{
		URL:        url,
		FinalURL:   finalURL,
		FoundLinks: discoveredLinks,
		StatusCode: resp.StatusCode,
		Duration:   duration,
	}
}
