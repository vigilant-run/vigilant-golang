package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// batcher is a struct that contains the queues for the logs, errors, metrics, and alerts
// it also contains the http client and the wait group
// when a batch is ready, the batcher will send it to the server
type batcher struct {
	token    string
	endpoint string

	logQueue    chan *logMessage
	errorQueue  chan *errorMessage
	metricQueue chan *metricMessage
	alertQueue  chan *alertMessage

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
		token:       token,
		endpoint:    endpoint,
		logQueue:    make(chan *logMessage, 1000),
		errorQueue:  make(chan *errorMessage, 1000),
		metricQueue: make(chan *metricMessage, 1000),
		alertQueue:  make(chan *alertMessage, 1000),
		batchStop:   make(chan struct{}),
		client:      httpClient,
	}
}

// start starts the batcher
func (b *batcher) start() {
	b.wg.Add(4)
	go b.runLogBatcher()
	go b.runErrorBatcher()
	go b.runMetricBatcher()
	go b.runAlertBatcher()
}

// addLog adds a log to the batcher's queue
func (b *batcher) addLog(message *logMessage) {
	if message == nil {
		return
	}
	b.logQueue <- message
}

// addError adds an error to the batcher's queue
func (b *batcher) addError(message *errorMessage) {
	if message == nil {
		return
	}
	b.errorQueue <- message
}

// addMetric adds a metric to the batcher's queue
func (b *batcher) addMetric(message *metricMessage) {
	if message == nil {
		return
	}
	b.metricQueue <- message
}

// addAlert adds an alert to the batcher's queue
func (b *batcher) addAlert(message *alertMessage) {
	if message == nil {
		return
	}
	b.alertQueue <- message
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

// runErrorBatcher runs the error batcher
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

// runAlertBatcher runs the alert batcher
func (b *batcher) runAlertBatcher() {
	defer b.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var alerts []*alertMessage

	for {
		select {
		case <-b.batchStop:
			if len(alerts) > 0 {
				b.sendAlertBatch(alerts)
			}
			return
		case msg := <-b.alertQueue:
			if msg == nil {
				continue
			}
			alerts = append(alerts, msg)
			if len(alerts) >= maxBatchSize {
				b.sendAlertBatch(alerts)
				alerts = nil
			}
		case <-ticker.C:
			if len(alerts) > 0 {
				b.sendAlertBatch(alerts)
				alerts = nil
			}
		}
	}
}

// runMetricBatcher runs the metric batcher
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

// sendErrorBatch sends an error batch to the server
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

// sendAlertBatch sends an alert batch to the server
func (b *batcher) sendAlertBatch(alerts []*alertMessage) {
	if len(alerts) == 0 {
		return
	}

	batch := &messageBatch{
		Token:  b.token,
		Type:   messageTypeAlert,
		Alerts: alerts,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return
	}

	err = b.sendBatch(batchBytes)
	if err != nil {
		fmt.Printf("error sending alert batch to %s: %v\n", b.endpoint, err)
	}
}

// sendMetricBatch sends a metric batch to the server
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
