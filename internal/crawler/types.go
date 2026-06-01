package crawler

import "time"

type Job struct {
	URL   string
	Depth int
}

type Result struct {
	URL        string
	FoundLinks []string
	StatusCode int
	Duration   time.Duration
	Err        error
}
