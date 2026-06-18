# Concurrent Web Crawler CLI

A highly efficient, concurrent web crawler written in Go. The application leverages a distributed worker pool architecture to fetch pages, extract links recursively up to a configurable depth, and streams structured data onto disk in real-time using a thread-safe JSON Lines (`.jsonl`) background worker.

## 🚀 Features

* **High Performance Concurrency:** Distributed worker pool pattern scales network requests effortlessly.
* **Graceful Shutdown:** Catches OS interrupts (`Ctrl+C`) to finish outstanding tasks and flush logs cleanly without data loss.
* **Robust HTML Parser:** Robust depth-first search (DFS) link resolution that automatically handles absolute/relative links, strips fragments, and deduplicates links page-by-page.
* **JSON Lines Storage Engine:** Decoupled background logging routine utilizes non-blocking Go channels to pipe and serialize results with zero thread contention or file corruption.
* **Smart Network Handling:** Custom User-Agent configurations and strict domain boundary locks to keep the crawler scoped to your seed hosts.

---

## 🛠️ Installation

Ensure you have Go installed (version 1.18+ recommended).

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/concurrent-web-crawler.git
   cd concurrent-web-crawler
   ```

2. Download dependencies (if any):
   ```
   go mod tidy
   ```

3. Build the executable:
   ```
   go build -o crawler ./cmd/crawler
   ```

---

## 📖 Usage & CLI Configurations

The crawler requires a seed list of URLs to begin. All other parameters fall back to safe, tested defaults.

`./crawler -urls="https://go.dev"`

### Command Line Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-urls` | string | *Required* | Comma-separated list of seed URLs to initiate the crawl. |
| `-file` | string | "crawl_output.txt" | Target output file name for storing results. |
| `-depth` | int | 2 | Maximum depth boundary for recursive link extraction. |
| `-workers`| int | 3 | Number of concurrent network worker routines to spin up. |
| `-timeout`| int | 5 | Maximum HTTP request timeout threshold in seconds. |
| `-exclude`| string | "" | Maximum HTTP request timeout threshold in seconds. |

### Execution Example

`./crawler -urls="https://go.dev" -workers=5 -depth=3 -file="my_crawl.jsonl" -timeout=10 -exclude="/docs/,/search/`

---

## 📊 Interpreting Output Data

The system outputs data using the **JSON Lines (`.jsonl`)** standard. Each row represents an isolated, valid JSON object corresponding to a crawled URL.

### Successful Fetch
```
{
  "url": "https://go.dev",
  "status_code": 200,
  "duration": "375.31075ms",
  "depth": 1,
  "found_links": [
    "https://go.dev/solutions/case-studies",
    "https://go.dev/learn/"
  ]
}
```

### Connection Error / Timeout
```
{
  "url": "https://some-broken-domain.com",
  "status_code": 0,
  "duration": "5.002s",
  "depth": 2,
  "error": "context deadline exceeded"
}
```

*Note: Successful fields that are empty (like `error` on success or `found_links` on a failure) are automatically omitted using `omitempty` to maintain an incredibly light storage footprint.*

---

## 🧪 Running Tests

The project separates core logic packages to maintain clean white-box and black-box testing boundaries. You can execute the test suite across parsing algorithms, mocks, and concurrent data pipelines via the Go toolchain.

Run all unit tests:
`go test ./...`

Run tests with the **Go Race Detector** enabled to check for concurrent pipeline memory leaks or thread conflicts:
`go test -race ./...`
