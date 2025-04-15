package vigilant

import (
	"log"
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
	log.Println("Starting metric collector...")
	c.wg.Add(2)
	go c.sender.start()
	go c.processEvents()
	go c.runTicker()
	log.Println("Metric collector started.")
}

// stop stops the collector and the sender using simplified shutdown logic
func (c *metricCollector) stop() {
	log.Println("Stopping metric collector...")
	close(c.stopChan)
	log.Println("Waiting for collector goroutines to finish...")
	c.wg.Wait()
	log.Println("Collector goroutines finished.")

	log.Println("Closing event channels...")
	close(c.counterEvents)
	close(c.gaugeEvents)
	log.Println("Event channels closed.")

	log.Println("Processing remaining events after shutdown...")
	c.processAfterShutdown()
	log.Println("Finished processing remaining events.")

	log.Println("Sending remaining metrics after shutdown...")
	c.sendAfterShutdown()
	log.Println("Finished sending remaining metrics.")

	c.sender.stop()
	log.Println("Metric collector stopped.")
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
	log.Println("Starting collector ticker...")
	now := time.Now()
	nextInterval := now.Truncate(c.interval).Add(c.interval)
	firstTriggerTime := nextInterval.Add(1 * time.Second)

	if firstTriggerTime.Before(now) {
		log.Printf("Adjusting first trigger time from %v", firstTriggerTime)
		firstTriggerTime = nextInterval.Add(c.interval).Add(1 * time.Second)
	}

	durationUntilFirstTrigger := firstTriggerTime.Sub(now)
	log.Printf("First tick scheduled for %v (in %v)", firstTriggerTime, durationUntilFirstTrigger)
	timer := time.NewTimer(durationUntilFirstTrigger)
	defer timer.Stop()

	var ticker *time.Ticker
	defer func() {
		if ticker != nil {
			ticker.Stop()
			log.Println("Ticker stopped.")
		}
	}()

	for {
		select {
		case <-c.stopChan:
			log.Println("Ticker received stop signal. Exiting.")
			return
		case firstTickTime := <-timer.C:
			intervalToProcess := firstTickTime.Truncate(c.interval).Add(-c.interval)
			log.Printf("Initial timer ticked at %v. Processing interval %v", firstTickTime, intervalToProcess)
			c.sendMetricsForInterval(intervalToProcess)

			log.Printf("Starting periodic ticker with interval %v", c.interval)
			ticker = time.NewTicker(c.interval)
			for {
				select {
				case <-c.stopChan:
					log.Println("Periodic ticker received stop signal during inner loop. Exiting.")
					return
				case tickTime := <-ticker.C:
					intervalToProcess = tickTime.Truncate(c.interval).Add(-c.interval)
					log.Printf("Periodic ticker ticked at %v. Processing interval %v", tickTime, intervalToProcess)
					c.sendMetricsForInterval(intervalToProcess)
				}
			}
		}
	}
}

// processEvents reads metric events from the channel and updates the buckets.
func (c *metricCollector) processEvents() {
	log.Println("Starting event processor...")
	defer c.wg.Done()
	for {
		select {
		case <-c.stopChan:
			log.Println("Event processor received stop signal. Exiting.")
			return
		case event, ok := <-c.counterEvents:
			if !ok {
				log.Println("Counter events channel closed.")
				continue
			}
			if event == nil {
				log.Println("Received nil counter event.")
				continue
			}
			c.processCounterEvent(event)
		case event, ok := <-c.gaugeEvents:
			if !ok {
				log.Println("Gauge events channel closed.")
				continue
			}
			if event == nil {
				log.Println("Received nil gauge event.")
				continue
			}
			c.processGaugeEvent(event)
		}
	}
}

// processAfterShutdown drains event channels after goroutines have stopped.
func (c *metricCollector) processAfterShutdown() {
	log.Println("Processing remaining counter events...")
	processedCounters := 0
	for event := range c.counterEvents {
		c.processCounterEvent(event)
		processedCounters++
	}
	log.Printf("Processed %d remaining counter events.", processedCounters)

	log.Println("Processing remaining gauge events...")
	processedGauges := 0
	for event := range c.gaugeEvents {
		c.processGaugeEvent(event)
		processedGauges++
	}
	log.Printf("Processed %d remaining gauge events.", processedGauges)
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
func (c *metricCollector) sendMetricsForInterval(intervalStart time.Time) {
	log.Printf("Attempting to send metrics for interval starting %v", intervalStart)

	var metricsToSend *aggregatedMetrics
	var counterCount, gaugeCount int
	processedBucket := false // Track if we found and processed the bucket

	c.mux.Lock()
	bucket, ok := c.capturedBuckets[intervalStart]
	if ok && bucket != nil {
		processedBucket = true // Mark as processed
		// Only aggregate if there's data
		if len(bucket.counters) > 0 || len(bucket.gauges) > 0 {
			log.Printf("Aggregating metrics for interval %v.", intervalStart)
			metricsToSend = aggregateCapturedMetrics(intervalStart, bucket)
			counterCount = len(metricsToSend.counterMetrics)
			gaugeCount = len(metricsToSend.gaugeMetrics)

			// Clear the bucket now that we've aggregated it.
			// Late metrics for this interval can still land here via getBucket,
			// but they won't be sent in this cycle.
			log.Printf("Clearing processed bucket for interval %v.", intervalStart)
			bucket.counters = make(map[string]*capturedCounter)
			bucket.gauges = make(map[string]*capturedGauge)
		} else {
			log.Printf("Bucket for interval %v exists but is empty.", intervalStart)
		}
	} else {
		log.Printf("No bucket found for interval %v. It might have been cleaned up or never created.", intervalStart)
	}
	c.mux.Unlock() // Unlock after potential aggregation and clearing

	if metricsToSend != nil && (counterCount > 0 || gaugeCount > 0) {
		log.Printf("Sending %d counters and %d gauges for interval %v.", counterCount, gaugeCount, intervalStart)
		c.sender.sendAggregatedMetrics(metricsToSend)
	} else if processedBucket {
		log.Printf("No metrics aggregated to send for interval %v (bucket was empty or cleared).", intervalStart)
	} else {
		log.Printf("No metrics bucket processed or sent for interval %v.", intervalStart)
	}

	// Cleanup based on the interval we just processed.
	// This will delete buckets older than intervalStart - c.interval
	c.cleanupOldBuckets(intervalStart)
}

// cleanupOldBuckets removes buckets older than the previous interval being processed.
// This gives late metrics potentially one extra interval to arrive before their bucket is deleted.
func (c *metricCollector) cleanupOldBuckets(currentIntervalJustProcessed time.Time) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Delete buckets whose start time is strictly before the *previous* interval.
	// e.g., if we just processed 18:55:00, the threshold is 18:54:00. Buckets < 18:54:00 are deleted.
	// The bucket for 18:54:00 and the just-cleared 18:55:00 bucket remain.
	cleanupThreshold := currentIntervalJustProcessed.Add(-1 * c.interval)
	log.Printf("Cleaning up buckets older than %v", cleanupThreshold)

	toDelete := []time.Time{}
	// Check only keys first to avoid holding lock while iterating potentially large buckets
	for ts := range c.capturedBuckets {
		if ts.Before(cleanupThreshold) {
			toDelete = append(toDelete, ts)
		}
	}

	if len(toDelete) > 0 {
		log.Printf("Cleaning up %d old buckets (older than %v): %v", len(toDelete), cleanupThreshold, toDelete)
		for _, ts := range toDelete {
			// Log if deleting a non-empty bucket (containing late/stranded metrics)
			bucketToDelete := c.capturedBuckets[ts]
			if bucketToDelete != nil && (len(bucketToDelete.counters) > 0 || len(bucketToDelete.gauges) > 0) {
				log.Printf("Warning: Deleting old bucket %v which contained %d counters and %d gauges (likely late metrics).",
					ts, len(bucketToDelete.counters), len(bucketToDelete.gauges))
			}
			delete(c.capturedBuckets, ts)
		}
	} else {
		log.Printf("No buckets found older than %v to clean up.", cleanupThreshold)
	}
}

// sendAfterShutdown sends all metrics currently held in buckets.
func (c *metricCollector) sendAfterShutdown() {
	log.Println("Attempting to send remaining metrics after shutdown...")

	bucketsToSend := make(map[time.Time]*capturedMetrics)

	c.mux.Lock()
	for ts, bucket := range c.capturedBuckets {
		if bucket != nil && (len(bucket.counters) > 0 || len(bucket.gauges) > 0) {
			bucketsToSend[ts] = bucket
		}
	}
	c.capturedBuckets = make(map[time.Time]*capturedMetrics)
	c.mux.Unlock()

	log.Printf("Found %d buckets with data to send after shutdown.", len(bucketsToSend))

	for timestamp, bucket := range bucketsToSend {
		log.Printf("Aggregating and sending metrics for shutdown bucket %v.", timestamp)
		aggregatedMetrics := aggregateCapturedMetrics(timestamp, bucket)
		log.Printf("Sending %d counters and %d gauges for shutdown interval %v.", len(aggregatedMetrics.counterMetrics), len(aggregatedMetrics.gaugeMetrics), timestamp)
		c.sender.sendAggregatedMetrics(aggregatedMetrics)
	}
	log.Println("Finished sending remaining metrics after shutdown.")
}

// getBucket gets the bucket for the current time
func (c *metricCollector) getBucket(now time.Time) *capturedMetrics {
	floored := now.Truncate(c.interval)
	bucket, ok := c.capturedBuckets[floored]
	if !ok {
		log.Printf("Creating new metrics bucket for interval %v", floored)
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
