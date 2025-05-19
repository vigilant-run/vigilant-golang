package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	logEndpoint     = "/api/message"
	maxLogBatchSize = 100
)

// logBatcher is a struct that contains the queues for the logs
// it also contains the http client and the wait group
// when a batch is ready, the logBatcher will send it to the server
type logBatcher struct {
	token    string
	endpoint string

	logQueue chan *logMessage

	client *http.Client

	stopped   bool
	batchStop chan struct{}
	wg        sync.WaitGroup
}

// newLogBatcher creates a new logBatcher
func newLogBatcher(
	token string,
	endpoint string,
	httpClient *http.Client,
) *logBatcher {
	return &logBatcher{
		token:     token,
		endpoint:  endpoint,
		logQueue:  make(chan *logMessage, 1000),
		batchStop: make(chan struct{}),
		client:    httpClient,
	}
}

// start starts the batcher
func (b *logBatcher) start() {
	b.wg.Add(1)
	go b.runLogBatcher()
}

// addLog adds a log to the batcher's queue
func (b *logBatcher) addLog(message *logMessage) {
	if message == nil || b.stopped {
		return
	}
	b.logQueue <- message
}

// stop stops the batcher and processes remaining logs
func (b *logBatcher) stop() {
	b.stopped = true
	close(b.batchStop)
	b.wg.Wait()

	close(b.logQueue)
	b.processAfterShutdown()
}

// runLogBatcher runs the log batcher
func (b *logBatcher) runLogBatcher() {
	defer b.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var logs []*logMessage
	for {
		select {
		case <-b.batchStop:
			if len(logs) > 0 {
				if err := b.sendLogBatch(logs); err != nil {
					fmt.Printf("error sending final log batch: %v\n", err)
				}
			}
			return
		case msg := <-b.logQueue:
			if msg == nil {
				continue
			}
			logs = append(logs, msg)
			if len(logs) >= maxLogBatchSize {
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

// processAfterShutdown processes any remaining logs in the queue after shutdown.
func (b *logBatcher) processAfterShutdown() {
	var logs []*logMessage
	for msg := range b.logQueue {
		if msg == nil {
			continue
		}
		logs = append(logs, msg)
		if len(logs) >= 100 {
			if err := b.sendLogBatch(logs); err != nil {
				fmt.Printf("error sending shutdown log batch: %v\n", err)
			}
			logs = nil
		}
	}
	if len(logs) > 0 {
		if err := b.sendLogBatch(logs); err != nil {
			fmt.Printf("error sending final shutdown log batch: %v\n", err)
		}
	}
}

// sendLogBatch sends a log batch to the server
func (b *logBatcher) sendLogBatch(logs []*logMessage) error {
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
func (b *logBatcher) sendBatch(batchBytes []byte) error {
	req, err := http.NewRequest("POST", b.endpoint+logEndpoint, bytes.NewBuffer(batchBytes))
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
