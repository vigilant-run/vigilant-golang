package vigilant

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// metricCollector is a struct that contains the buckets for the metrics
// it also contains the sender for the metrics
// every interval, the collector will send the client-side aggregated metrics to the server
type metricCollector struct {
	interval        time.Duration
	capturedBuckets map[time.Time]*capturedMetrics
	sender          *metricSender
	counterEvents   chan *counterEvent
	gaugeEvents     chan *gaugeEvent

	mux      sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup // Use WaitGroup for shutdown coordination
}

// newMetricCollector creates a new metricCollector
func newMetricCollector(
	interval time.Duration,
	token string,
	endpoint string,
	httpClient *http.Client,
) *metricCollector {
	metricSender := newMetricSender(
		token,
		endpoint,
		httpClient,
	)
	return &metricCollector{
		interval:        interval,
		capturedBuckets: make(map[time.Time]*capturedMetrics),
		sender:          metricSender,
		counterEvents:   make(chan *counterEvent, 1000),
		gaugeEvents:     make(chan *gaugeEvent, 1000),

		mux:      sync.RWMutex{},
		stopChan: make(chan struct{}),
		// wg is implicitly initialized
	}
}

// start starts the collector, the sender, and the event processor
func (c *metricCollector) start() {
	c.wg.Add(2)
	go c.sender.start()
	go c.processEvents()
	go c.runTicker()
}

// stop stops the collector and the sender using simplified shutdown logic
func (c *metricCollector) stop() {
	close(c.stopChan)
	c.wg.Wait()

	close(c.counterEvents)
	close(c.gaugeEvents)

	c.processAfterShutdown()
	c.sendAfterShutdown()
	c.sender.stop()
}

// addCounter adds a counter event to the collector
func (c *metricCollector) addCounter(event *counterEvent) {
	c.counterEvents <- event
}

// addGauge adds a gauge event to the collector
func (c *metricCollector) addGauge(event *gaugeEvent) {
	c.gaugeEvents <- event
}

// runTicker runs the ticker for the collector
func (c *metricCollector) runTicker() {
	defer c.wg.Done()

	now := time.Now()
	nextInterval := now.Truncate(c.interval).Add(c.interval)
	firstTriggerTime := nextInterval.Add(1 * time.Second)

	if firstTriggerTime.Before(now) {
		firstTriggerTime = nextInterval.Add(c.interval).Add(1 * time.Second)
	}

	durationUntilFirstTrigger := firstTriggerTime.Sub(now)
	timer := time.NewTimer(durationUntilFirstTrigger)
	defer timer.Stop()

	var ticker *time.Ticker
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	for {
		select {
		case <-c.stopChan:
			return
		case firstTickTime := <-timer.C:
			c.sendMetricsForInterval(firstTickTime.Truncate(c.interval))
			ticker = time.NewTicker(c.interval)
			for {
				select {
				case <-c.stopChan:
					return
				case tickTime := <-ticker.C:
					c.sendMetricsForInterval(tickTime.Truncate(c.interval))
				}
			}
		}
	}
}

// processEvents reads metric events from the channel and updates the buckets.
func (c *metricCollector) processEvents() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stopChan:
			return
		case event := <-c.counterEvents:
			if event == nil {
				continue
			}
			c.processCounterEvent(event)
		case event := <-c.gaugeEvents:
			if event == nil {
				continue
			}
			c.processGaugeEvent(event)
		}
	}
}

// processAfterShutdown drains event channels after goroutines have stopped.
func (c *metricCollector) processAfterShutdown() {
	for event := range c.counterEvents {
		c.processCounterEvent(event)
	}
	for event := range c.gaugeEvents {
		c.processGaugeEvent(event)
	}
}

// processCounterEvent handles processing a single counter event
func (c *metricCollector) processCounterEvent(event *counterEvent) {
	c.mux.Lock()
	defer c.mux.Unlock()

	bucket := c.getBucket(event.timestamp)

	identifier := newMetricIdentifier(event.name, event.tags)
	identifierString := identifier.String()

	if counter, exists := bucket.counters[identifierString]; exists {
		counter.value += event.value
	} else {
		bucket.counters[identifierString] = &capturedCounter{
			name:  event.name,
			value: event.value,
			tags:  event.tags,
		}
	}
}

// processGaugeEvent handles processing a single gauge event
func (c *metricCollector) processGaugeEvent(event *gaugeEvent) {
	c.mux.Lock()
	defer c.mux.Unlock()

	bucket := c.getBucket(event.timestamp)

	identifier := newMetricIdentifier(event.name, event.tags)
	identifierString := identifier.String()

	if gauge, exists := bucket.gauges[identifierString]; exists {
		gauge.value = event.value
	} else {
		bucket.gauges[identifierString] = &capturedGauge{
			name:  event.name,
			value: event.value,
			tags:  event.tags,
		}
	}
}

// sendMetricsForInterval sends the metrics for the interval
func (c *metricCollector) sendMetricsForInterval(now time.Time) {
	c.mux.Lock()
	defer c.mux.Unlock()

	intervalStart := now.Truncate(c.interval)
	bucket, ok := c.capturedBuckets[intervalStart]
	if !ok {
		return
	}

	aggregatedMetrics := aggregateCapturedMetrics(intervalStart, bucket)
	delete(c.capturedBuckets, intervalStart)

	c.sender.sendAggregatedMetrics(aggregatedMetrics)
}

// sendAfterShutdown sends all metrics currently held in buckets.
func (c *metricCollector) sendAfterShutdown() {
	c.mux.Lock()
	defer c.mux.Unlock()

	bucketTimestamps := make([]time.Time, 0, len(c.capturedBuckets))
	for ts := range c.capturedBuckets {
		bucketTimestamps = append(bucketTimestamps, ts)
	}

	for _, timestamp := range bucketTimestamps {
		bucket := c.capturedBuckets[timestamp]
		if bucket != nil {
			aggregatedMetrics := aggregateCapturedMetrics(timestamp, bucket)
			c.sender.sendAggregatedMetrics(aggregatedMetrics)
			delete(c.capturedBuckets, timestamp)
		}
	}
}

// getBucket gets the bucket for the current time
func (c *metricCollector) getBucket(now time.Time) *capturedMetrics {
	floored := now.Truncate(c.interval)
	bucket, ok := c.capturedBuckets[floored]
	if !ok {
		bucket = createCapturedMetrics()
		c.capturedBuckets[floored] = bucket
	}
	return bucket
}

// metricIdentifier is a struct that contains the name and tags of a metric
type metricIdentifier struct {
	name string
	tags map[string]string
}

func newMetricIdentifier(name string, tags map[string]string) *metricIdentifier {
	return &metricIdentifier{name: name, tags: tags}
}

// String returns the string representation of the metric identifier
func (m *metricIdentifier) String() string {
	parts := []string{m.name}
	for k, v := range m.tags {
		parts = append(parts, k, v)
	}
	return strings.Join(parts, "_")
}

// capturedCounter is a struct that contains the name, value, and tags of a counter metric
type capturedCounter struct {
	name  string
	value float64
	tags  map[string]string
}

// capturedGauge is a struct that contains the name, value, and tags of a gauge metric
type capturedGauge struct {
	name  string
	value float64
	tags  map[string]string
}

// capturedMetrics is a struct that contains the counters and gauges for a bucket
type capturedMetrics struct {
	counters map[string]*capturedCounter
	gauges   map[string]*capturedGauge
}

// createCapturedMetrics creates a new capturedMetrics
func createCapturedMetrics() *capturedMetrics {
	return &capturedMetrics{
		counters: make(map[string]*capturedCounter),
		gauges:   make(map[string]*capturedGauge),
	}
}

// transformCapturedMetrics transforms the captured metrics into an aggregated metrics
func aggregateCapturedMetrics(
	timestamp time.Time,
	capturedMetrics *capturedMetrics,
) *aggregatedMetrics {
	aggregatedMetrics := newAggregatedMetrics()

	for _, counter := range capturedMetrics.counters {
		aggregatedMetrics.counterMetrics = append(aggregatedMetrics.counterMetrics, &counterMessage{
			Timestamp:  timestamp,
			MetricName: counter.name,
			Value:      counter.value,
			Tags:       counter.tags,
		})
	}

	for _, gauge := range capturedMetrics.gauges {
		aggregatedMetrics.gaugeMetrics = append(aggregatedMetrics.gaugeMetrics, &gaugeMessage{
			Timestamp:  timestamp,
			MetricName: gauge.name,
			Value:      gauge.value,
			Tags:       gauge.tags,
		})
	}

	return aggregatedMetrics
}
