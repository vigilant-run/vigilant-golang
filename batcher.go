package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type batcher struct {
	token    string
	endpoint string

	logQueue    chan *logMessage
	errorQueue  chan *errorMessage
	metricQueue chan *metricMessage

	client *http.Client

	batchStop chan struct{}
	wg        sync.WaitGroup
}

func newBatcher(
	token string,
	endpoint string,
	httpClient *http.Client,
) *batcher {
	return &batcher{
		token:       token,
		endpoint:    endpoint,
		logQueue:    make(chan *logMessage, 1000),
		errorQueue:  make(chan *errorMessage, 1000),
		metricQueue: make(chan *metricMessage, 1000),
		batchStop:   make(chan struct{}),
		client:      httpClient,
	}
}

func (b *batcher) start() {
	b.wg.Add(3)
	go b.runLogBatcher()
	go b.runErrorBatcher()
	go b.runMetricBatcher()
}

func (b *batcher) addLog(message *logMessage) {
	if message == nil {
		return
	}
	b.logQueue <- message
}

func (b *batcher) addError(message *errorMessage) {
	if message == nil {
		return
	}
	b.errorQueue <- message
}

func (b *batcher) addMetric(message *metricMessage) {
	if message == nil {
		return
	}
	b.metricQueue <- message
}

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

func (b *batcher) runErrorBatcher() {
	defer b.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var errors []*errorMessage

	for {
		select {
		case <-b.batchStop:
			if len(errors) > 0 {
				b.sendErrorBatch(errors)
			}
			return
		case msg := <-b.errorQueue:
			if msg == nil {
				continue
			}
			errors = append(errors, msg)
			if len(errors) >= maxBatchSize {
				b.sendErrorBatch(errors)
				errors = nil
			}
		case <-ticker.C:
			if len(errors) > 0 {
				b.sendErrorBatch(errors)
				errors = nil
			}
		}
	}
}

func (b *batcher) runMetricBatcher() {
	defer b.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var metrics []*metricMessage

	for {
		select {
		case <-b.batchStop:
			if len(metrics) > 0 {
				b.sendMetricBatch(metrics)
			}
			return
		case msg := <-b.metricQueue:
			if msg == nil {
				continue
			}
			metrics = append(metrics, msg)
			if len(metrics) >= maxBatchSize {
				b.sendMetricBatch(metrics)
				metrics = nil
			}
		case <-ticker.C:
			if len(metrics) > 0 {
				b.sendMetricBatch(metrics)
				metrics = nil
			}
		}
	}
}

func (b *batcher) stop() {
	close(b.batchStop)
	b.wg.Wait()
}

func (b *batcher) sendLogBatch(logs []*logMessage) error {
	if len(logs) == 0 {
		return nil
	}

	batch := &messageBatch{
		Token: b.token,
		Type:  messageTypeLog,
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

func (b *batcher) sendErrorBatch(errors []*errorMessage) {
	if len(errors) == 0 {
		return
	}

	batch := &messageBatch{
		Token:  b.token,
		Type:   messageTypeError,
		Errors: errors,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return
	}

	err = b.sendBatch(batchBytes)
	if err != nil {
		fmt.Printf("error sending error batch to %s: %v\n", b.endpoint, err)
	}
}

func (b *batcher) sendMetricBatch(metrics []*metricMessage) {
	if len(metrics) == 0 {
		return
	}

	batch := &messageBatch{
		Token:   b.token,
		Type:    messageTypeMetric,
		Metrics: metrics,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return
	}

	err = b.sendBatch(batchBytes)
	if err != nil {
		fmt.Printf("error sending metric batch to %s: %v\n", b.endpoint, err)
	}
}

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
