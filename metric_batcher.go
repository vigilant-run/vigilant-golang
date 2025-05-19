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
	maxMetricBatchSize = 100
)

// metricBatcher is a struct that contains the queues for the metrics
// it also contains the http client and the wait group
// when a batch is ready, the metricBatcher will send it to the server
type metricBatcher struct {
	token    string
	endpoint string

	metricQueue chan *metricMessage

	client *http.Client

	stopped   bool
	batchStop chan struct{}
	wg        sync.WaitGroup
}

// newMetricBatcher creates a new metricBatcher
func newMetricBatcher(
	token string,
	endpoint string,
	httpClient *http.Client,
) *metricBatcher {
	return &metricBatcher{
		token:       token,
		endpoint:    endpoint,
		metricQueue: make(chan *metricMessage, 1000),
		batchStop:   make(chan struct{}),
		client:      httpClient,
	}
}

// start starts the batcher
func (b *metricBatcher) start() {
	b.wg.Add(1)
	go b.runMetricBatcher()
}

// addMetric adds a metric to the batcher's queue
func (b *metricBatcher) addMetric(message *metricMessage) {
	if message == nil || b.stopped {
		return
	}
	b.metricQueue <- message
}

// stop stops the batcher and processes remaining metrics
func (b *metricBatcher) stop() {
	b.stopped = true
	close(b.batchStop)
	b.wg.Wait()

	close(b.metricQueue)
	b.processAfterShutdown()
}

// runMetricBatcher runs the metric batcher
func (b *metricBatcher) runMetricBatcher() {
	defer b.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var metrics []*metricMessage
	for {
		select {
		case <-b.batchStop:
			if len(metrics) > 0 {
				if err := b.sendMetricBatch(metrics); err != nil {
					fmt.Printf("error sending final metric batch: %v\n", err)
				}
			}
			return
		case msg := <-b.metricQueue:
			if msg == nil {
				continue
			}
			metrics = append(metrics, msg)
			if len(metrics) >= maxMetricBatchSize {
				if err := b.sendMetricBatch(metrics); err != nil {
					fmt.Printf("error sending metric batch: %v\n", err)
				}
				metrics = nil
			}
		case <-ticker.C:
			if len(metrics) > 0 {
				if err := b.sendMetricBatch(metrics); err != nil {
					fmt.Printf("error sending metric batch: %v\n", err)
				}
				metrics = nil
			}
		}
	}
}

// processAfterShutdown processes any remaining metrics in the queue after shutdown.
func (b *metricBatcher) processAfterShutdown() {
	var metrics []*metricMessage
	for msg := range b.metricQueue {
		if msg == nil {
			continue
		}
		metrics = append(metrics, msg)
		if len(metrics) >= 100 {
			if err := b.sendMetricBatch(metrics); err != nil {
				fmt.Printf("error sending shutdown metric batch: %v\n", err)
			}
			metrics = nil
		}
	}
	if len(metrics) > 0 {
		if err := b.sendMetricBatch(metrics); err != nil {
			fmt.Printf("error sending final shutdown metric batch: %v\n", err)
		}
	}
}

// sendMetricBatch sends a metric batch to the server
func (b *metricBatcher) sendMetricBatch(metrics []*metricMessage) error {
	if len(metrics) == 0 {
		return nil
	}

	batch := &messageBatch{
		Token:   b.token,
		Metrics: metrics,
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
func (b *metricBatcher) sendBatch(batchBytes []byte) error {
	req, err := http.NewRequest("POST", b.endpoint+metricEndpoint, bytes.NewBuffer(batchBytes))
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
