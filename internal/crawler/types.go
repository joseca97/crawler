package crawler

import (
	"encoding/json"
	"time"
)

type Job struct {
	URL   string
	Depth int
}

type Result struct {
	URL        string
	FinalURL   string
	FoundLinks []string
	StatusCode int
	Duration   time.Duration
	Err        error
	Depth      int
}

func (r Result) MarshalJSON() ([]byte, error) {
	var errString string
	if r.Err != nil {
		errString = r.Err.Error()
	}

	return json.Marshal(&struct {
		Duration   string   `json:"duration"`
		Err        string   `json:"error,omitempty"`
		URL        string   `json:"url"`
		FinalURL   string   `json:"final_url,omitempty"`
		StatusCode int      `json:"status_code"`
		FoundLinks []string `json:"found_links,omitempty"`
		Depth      int      `json:"depth"`
	}{
		Duration:   r.Duration.String(),
		Err:        errString,
		URL:        r.URL,
		FinalURL:   r.FinalURL,
		StatusCode: r.StatusCode,
		FoundLinks: r.FoundLinks,
		Depth:      r.Depth,
	})
}
