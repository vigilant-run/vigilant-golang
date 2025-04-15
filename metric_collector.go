package vigilant

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"
)

// metricCollector is a struct that contains the buckets for the metrics
// it also contains the sender for the metrics
// every interval, the collector will send the client-side aggregated metrics to the server
type metricCollector struct {
	sender       *metricSender
	registration *registrationHandler

	interval        time.Duration
	capturedBuckets map[time.Time]*capturedMetrics
	counterEvents   chan *metricEvent
	gaugeEvents     chan *metricEvent
	histogramEvents chan *metricEvent

	mux      sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// newMetricCollector creates a new metricCollector
func newMetricCollector(
	interval time.Duration,
	token string,
	endpoint string,
	serviceName string,
	httpClient *http.Client,
) *metricCollector {
	metricSender := newMetricSender(
		token,
		endpoint,
		httpClient,
	)
	registrationHandler := newRegistrationHandler(
		token,
		endpoint,
		serviceName,
		httpClient,
	)
	return &metricCollector{
		sender:          metricSender,
		registration:    registrationHandler,
		interval:        interval,
		capturedBuckets: make(map[time.Time]*capturedMetrics),
		counterEvents:   make(chan *metricEvent, 1000),
		gaugeEvents:     make(chan *metricEvent, 1000),
		histogramEvents: make(chan *metricEvent, 1000),
		mux:             sync.RWMutex{},
		stopChan:        make(chan struct{}),
		wg:              sync.WaitGroup{},
	}
}

// start starts the collector, the sender, and the event processor
func (c *metricCollector) start() {
	c.wg.Add(2)
	go c.sender.start()
	go c.registration.start()
	go c.processEvents()
	go c.runTicker()
}

// stop stops the collector and the sender using simplified shutdown logic
func (c *metricCollector) stop() {
	close(c.stopChan)
	c.wg.Wait()

	close(c.counterEvents)
	close(c.gaugeEvents)
	close(c.histogramEvents)
	c.processAfterShutdown()

	c.sendAfterShutdown()

	c.sender.stop()
	c.registration.stop()
}

// addCounter adds a counter event to the collector
func (c *metricCollector) addCounter(event *metricEvent) {
	c.counterEvents <- event
}

// addGauge adds a gauge event to the collector
func (c *metricCollector) addGauge(event *metricEvent) {
	c.gaugeEvents <- event
}

// addHistogram adds a histogram event to the collector
func (c *metricCollector) addHistogram(event *metricEvent) {
	c.histogramEvents <- event
}

// runTicker runs the ticker for the collector
func (c *metricCollector) runTicker() {
	defer c.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-c.stopChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	err := c.registration.waitForRegistration(ctx)
	if err != nil {
		fmt.Printf("metric collector exiting ticker: could not complete initial registration: %v\n", err)
		return
	}

	now := time.Now()
	nextInterval := now.Truncate(c.interval).Add(c.interval)
	firstTriggerTime := nextInterval.Add(1 * time.Second)

	if firstTriggerTime.Before(now) {
		firstTriggerTime = nextInterval.Add(c.interval).Add(1 * time.Second)
	}

	durationUntilFirstTrigger := firstTriggerTime.Sub(now)
	if durationUntilFirstTrigger < 0 {
		durationUntilFirstTrigger = 0
	}

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
			select {
			case <-c.stopChan:
				return
			default:
			}

			intervalToProcess := firstTickTime.Truncate(c.interval).Add(-c.interval)
			c.sendMetricsForInterval(intervalToProcess)

			ticker = time.NewTicker(c.interval)

			for {
				select {
				case <-c.stopChan:
					return
				case tickTime := <-ticker.C:
					intervalToProcess = tickTime.Truncate(c.interval).Add(-c.interval)
					c.sendMetricsForInterval(intervalToProcess)
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
		case event, ok := <-c.counterEvents:
			if !ok {
				continue
			}
			if event == nil {
				continue
			}
			c.processCounterEvent(event)
		case event, ok := <-c.gaugeEvents:
			if !ok {
				continue
			}
			if event == nil {
				continue
			}
			c.processGaugeEvent(event)
		case event, ok := <-c.histogramEvents:
			if !ok {
				continue
			}
			if event == nil {
				continue
			}
			c.processHistogramEvent(event)
		}
	}
}

// processCounterEvent handles processing a single counter event
func (c *metricCollector) processCounterEvent(event *metricEvent) {
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
			tags:  event.tags,
			value: event.value,
		}
	}
}

// processGaugeEvent handles processing a single gauge event
func (c *metricCollector) processGaugeEvent(event *metricEvent) {
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
			tags:  event.tags,
			value: event.value,
		}
	}
}

// processHistogramEvent handles processing a single histogram event
func (c *metricCollector) processHistogramEvent(event *metricEvent) {
	c.mux.Lock()
	defer c.mux.Unlock()

	bucket := c.getBucket(event.timestamp)

	identifier := newMetricIdentifier(event.name, event.tags)
	identifierString := identifier.String()

	if histogram, exists := bucket.histograms[identifierString]; exists {
		histogram.values = append(histogram.values, event.value)
	} else {
		bucket.histograms[identifierString] = &capturedHistogram{
			name:   event.name,
			values: []float64{event.value},
			tags:   event.tags,
		}
	}
}

// processAfterShutdown drains event channels after goroutines have stopped.
func (c *metricCollector) processAfterShutdown() {
	processedCounters := 0
	for event := range c.counterEvents {
		c.processCounterEvent(event)
		processedCounters++
	}

	processedGauges := 0
	for event := range c.gaugeEvents {
		c.processGaugeEvent(event)
		processedGauges++
	}
}

// sendMetricsForInterval sends the metrics for the interval
func (c *metricCollector) sendMetricsForInterval(intervalStart time.Time) {
	serviceInstance, err := c.registration.getServiceInstance()
	if err != nil {
		return
	}

	var metricsToSend *aggregatedMetrics
	var counterCount, gaugeCount int

	c.mux.Lock()
	bucket, ok := c.capturedBuckets[intervalStart]
	if ok && bucket != nil && (len(bucket.counters) > 0 || len(bucket.gauges) > 0) {
		metricsToSend = aggregateCapturedMetrics(intervalStart, bucket, serviceInstance)
		counterCount = len(metricsToSend.counterMetrics)
		gaugeCount = len(metricsToSend.gaugeMetrics)

		bucket.counters = make(map[string]*capturedCounter)
		bucket.gauges = make(map[string]*capturedGauge)
	}
	c.mux.Unlock()

	if metricsToSend != nil && (counterCount > 0 || gaugeCount > 0) {
		c.sender.sendAggregatedMetrics(metricsToSend)
	}

	c.cleanupOldBuckets(intervalStart)
}

// cleanupOldBuckets removes buckets older than the previous interval being processed.
// This gives late metrics potentially one extra interval to arrive before their bucket is deleted.
func (c *metricCollector) cleanupOldBuckets(currentIntervalJustProcessed time.Time) {
	c.mux.Lock()
	defer c.mux.Unlock()

	cleanupThreshold := currentIntervalJustProcessed.Add(-1 * c.interval)
	toDelete := []time.Time{}
	for ts := range c.capturedBuckets {
		if ts.Before(cleanupThreshold) {
			toDelete = append(toDelete, ts)
		}
	}

	if len(toDelete) > 0 {
		for _, ts := range toDelete {
			delete(c.capturedBuckets, ts)
		}
	}
}

// sendAfterShutdown sends all metrics currently held in buckets.
func (c *metricCollector) sendAfterShutdown() {
	bucketsToSend := make(map[time.Time]*capturedMetrics)

	c.mux.Lock()
	for ts, bucket := range c.capturedBuckets {
		if bucket != nil && (len(bucket.counters) > 0 || len(bucket.gauges) > 0) {
			bucketsToSend[ts] = bucket
		}
	}
	c.capturedBuckets = make(map[time.Time]*capturedMetrics)
	c.mux.Unlock()

	serviceInstance, err := c.registration.getServiceInstance()
	if err != nil {
		return
	}

	for timestamp, bucket := range bucketsToSend {
		aggregatedMetrics := aggregateCapturedMetrics(timestamp, bucket, serviceInstance)
		c.sender.sendAggregatedMetrics(aggregatedMetrics)
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

// capturedHistogram is a struct that contains the name, value, and tags of a histogram metric
type capturedHistogram struct {
	name   string
	values []float64
	tags   map[string]string
}

// capturedMetrics is a struct that contains the counters and gauges for a bucket
type capturedMetrics struct {
	counters   map[string]*capturedCounter
	gauges     map[string]*capturedGauge
	histograms map[string]*capturedHistogram
}

// createCapturedMetrics creates a new capturedMetrics
func createCapturedMetrics() *capturedMetrics {
	return &capturedMetrics{
		counters:   make(map[string]*capturedCounter),
		gauges:     make(map[string]*capturedGauge),
		histograms: make(map[string]*capturedHistogram),
	}
}

// transformCapturedMetrics transforms the captured metrics into an aggregated metrics
func aggregateCapturedMetrics(
	timestamp time.Time,
	capturedMetrics *capturedMetrics,
	serviceInstance string,
) *aggregatedMetrics {
	aggregatedMetrics := newAggregatedMetrics()

	for _, counter := range capturedMetrics.counters {
		aggregatedMetrics.counterMetrics = append(aggregatedMetrics.counterMetrics, &counterMessage{
			Timestamp:  timestamp,
			MetricName: counter.name,
			Value:      counter.value,
			Tags:       addServiceTag(counter.tags, serviceInstance),
		})
	}

	for _, gauge := range capturedMetrics.gauges {
		aggregatedMetrics.gaugeMetrics = append(aggregatedMetrics.gaugeMetrics, &gaugeMessage{
			Timestamp:  timestamp,
			MetricName: gauge.name,
			Value:      gauge.value,
			Tags:       addServiceTag(gauge.tags, serviceInstance),
		})
	}

	for _, histogram := range capturedMetrics.histograms {
		aggregatedMetrics.histogramMetrics = append(aggregatedMetrics.histogramMetrics, &histogramMessage{
			Timestamp:  timestamp,
			MetricName: histogram.name,
			Values:     histogram.values,
			Tags:       addServiceTag(histogram.tags, serviceInstance),
		})
	}

	return aggregatedMetrics
}

func addServiceTag(tags map[string]string, service string) map[string]string {
	finalTags := make(map[string]string, len(tags)+1)
	finalTags["service"] = service
	maps.Copy(finalTags, tags)
	return finalTags
}
