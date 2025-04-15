package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// batcher is a struct that contains the queues for the logs
// it also contains the http client and the wait group
// when a batch is ready, the batcher will send it to the server
type batcher struct {
	token    string
	endpoint string

	logQueue chan *logMessage

	client *http.Client

	batchStop chan struct{}
	wg        sync.WaitGroup
}

// newBatcher creates a new batcher
func newBatcher(
	token string,
	endpoint string,
	httpClient *http.Client,
) *batcher {
	return &batcher{
		token:     token,
		endpoint:  endpoint,
		logQueue:  make(chan *logMessage, 1000),
		batchStop: make(chan struct{}),
		client:    httpClient,
	}
}

// start starts the batcher
func (b *batcher) start() {
	b.wg.Add(1)
	go b.runLogBatcher()
}

// addLog adds a log to the batcher's queue
func (b *batcher) addLog(message *logMessage) {
	if message == nil {
		return
	}
	b.logQueue <- message
}

// runLogBatcher runs the log batcher
func (b *batcher) runLogBatcher() {
	defer b.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var logs []*logMessage
	for {
		select {
		case <-b.batchStop:
			if len(logs) > 0 {
				if err := b.sendLogBatch(logs); err != nil {
					fmt.Printf("error sending log batch: %v\n", err)
				}
			}
			return
		case msg := <-b.logQueue:
			if msg == nil {
				continue
			}
			logs = append(logs, msg)
			if len(logs) >= maxBatchSize {
				if err := b.sendLogBatch(logs); err != nil {
					fmt.Printf("error sending log batch: %v\n", err)
				}
				logs = nil
			}
		case <-ticker.C:
			if len(logs) > 0 {
				if err := b.sendLogBatch(logs); err != nil {
					fmt.Printf("error sending log batch: %v\n", err)
				}
				logs = nil
			}
		}
	}
}

// stop stops the batcher
func (b *batcher) stop() {
	close(b.batchStop)
	b.wg.Wait()
}

// sendLogBatch sends a log batch to the server
func (b *batcher) sendLogBatch(logs []*logMessage) error {
	if len(logs) == 0 {
		return nil
	}

	batch := &messageBatch{
		Token: b.token,
		Logs:  logs,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	err = b.sendBatch(batchBytes)
	if err != nil {
		return err
	}

	return nil
}

// sendBatch sends a batch to the server
func (b *batcher) sendBatch(batchBytes []byte) error {
	req, err := http.NewRequest("POST", b.endpoint, bytes.NewBuffer(batchBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.token)

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
