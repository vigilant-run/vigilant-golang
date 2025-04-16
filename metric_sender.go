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

	stopped   bool
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
		stopped:   false,
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
	if metrics == nil || s.stopped {
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
			if len(aggs.counterMetrics) > 0 || len(aggs.gaugeMetrics) > 0 || len(aggs.histogramMetrics) > 0 {
				s.sendMetrics(aggs)
			}
		}
	}
}

// stop stops the sender and processes remaining metrics
func (s *metricSender) stop() {
	s.stopped = true
	close(s.batchStop)
	s.wg.Wait()

	close(s.aggsQueue)
	s.processAfterShutdown()
}

// processAfterShutdown processes any remaining aggregated metrics in the queue after shutdown.
func (s *metricSender) processAfterShutdown() {
	for aggs := range s.aggsQueue {
		if aggs == nil {
			continue
		}
		if len(aggs.counterMetrics) > 0 || len(aggs.gaugeMetrics) > 0 || len(aggs.histogramMetrics) > 0 {
			_ = s.sendMetrics(aggs)
		}
	}
}

// sendMetrics sends a counter batch to the server
func (s *metricSender) sendMetrics(
	metrics *aggregatedMetrics,
) error {
	counterCount := len(metrics.counterMetrics)
	gaugeCount := len(metrics.gaugeMetrics)
	histogramCount := len(metrics.histogramMetrics)

	if counterCount == 0 && gaugeCount == 0 && histogramCount == 0 {
		return nil
	}

	batch := &messageBatch{
		Token: s.token,
	}

	batch.MetricsCounters = metrics.counterMetrics
	batch.MetricsGauges = metrics.gaugeMetrics
	batch.MetricsHistograms = metrics.histogramMetrics

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
