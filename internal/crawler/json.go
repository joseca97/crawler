package crawler

import (
	"fmt"
	"os"
)

type JSONLStorage struct {
	resultsChan chan Result
	doneChan    chan bool
	fileName    string
}

func NewJSONLStorage(filename string, bufferSize int) *JSONLStorage {
	s := &JSONLStorage{
		resultsChan: make(chan Result, bufferSize),
		doneChan:    make(chan bool),
		fileName:    filename,
	}

	go s.startWorker()
	return s
}

func (s *JSONLStorage) Write(res Result) {
	s.resultsChan <- res
}

func (s *JSONLStorage) Close() {
	close(s.resultsChan)
	<-s.doneChan
}

func (s *JSONLStorage) startWorker() {
	file, err := os.OpenFile(s.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to open storage file: %v\n", err)
		s.doneChan <- false
		return
	}
	defer file.Close()

	for result := range s.resultsChan {
		jsonData, err := result.MarshalJSON()
		if err != nil {
			fmt.Printf("[ERROR] Failed to marshal result for %s: %v\n", result.URL, err)
			continue
		}
		if _, err := file.Write(append(jsonData, '\n')); err != nil {
			fmt.Printf("[ERROR] Failed to write to file: %v\n", err)
		}
	}
	s.doneChan <- true
}
