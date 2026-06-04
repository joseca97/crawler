# Concurrent Web Crawler CLI

A highly efficient, concurrent web crawler written in Go. The application leverages a distributed worker pool architecture to fetch pages, extract links recursively up to a configurable depth, and streams structured data onto disk in real-time using a thread-safe JSON Lines (`.jsonl`) background worker.