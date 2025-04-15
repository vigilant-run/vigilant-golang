package vigilant

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

// metricSender is a struct that contains the queues for the metrics
// it immediately sends batches of metrics to the server
type metricSender struct {
	token    string
	endpoint string

	aggsQueue chan *aggregatedMetrics

	client *http.Client

	batchStop chan struct{}
	wg        sync.WaitGroup
}

// newMetricSender creates a new metricSender
func newMetricSender(
	token string,
	endpoint string,
	httpClient *http.Client,
) *metricSender {
	return &metricSender{
		token:     token,
		endpoint:  endpoint,
		aggsQueue: make(chan *aggregatedMetrics, 100),
		batchStop: make(chan struct{}),
		client:    httpClient,
	}
}

// start starts the sender
func (s *metricSender) start() {
	s.wg.Add(1)
	go s.runMetricSender()
}

// sendAggregatedMetrics sends a batch to the sender's queue
func (s *metricSender) sendAggregatedMetrics(metrics *aggregatedMetrics) {
	if metrics == nil {
		return
	}
	s.aggsQueue <- metrics
}

// runMetricSender runs the metric sender
func (s *metricSender) runMetricSender() {
	defer s.wg.Done()

	for {
		select {
		case <-s.batchStop:
			return
		case aggs := <-s.aggsQueue:
			if aggs == nil {
				continue
			}
			if len(aggs.counterMetrics) > 0 || len(aggs.gaugeMetrics) > 0 {
				s.sendMetrics(aggs)
			}
		}
	}
}

// stop stops the sender
func (s *metricSender) stop() {
	close(s.batchStop)
	s.wg.Wait()
}

// sendMetrics sends a counter batch to the server
func (s *metricSender) sendMetrics(
	metrics *aggregatedMetrics,
) error {
	counterCount := len(metrics.counterMetrics)
	gaugeCount := len(metrics.gaugeMetrics)

	if counterCount == 0 && gaugeCount == 0 {
		return nil
	}

	batch := &messageBatch{
		Token: s.token,
	}

	batch.MetricsCounters = metrics.counterMetrics
	batch.MetricsGauges = metrics.gaugeMetrics

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	err = s.sendBatch(batchBytes)
	if err != nil {
		return err
	}

	return nil
}

// sendBatch sends a batch to the server
func (s *metricSender) sendBatch(batchBytes []byte) error {
	req, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(batchBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
