package vigilant

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// metricCollector is a struct that contains the buckets for the metrics
// it also contains the sender for the metrics
// every interval, the collector will send the client-side aggregated metrics to the server
type metricCollector struct {
	sender *metricSender

	interval time.Duration

	counterSeries   map[string]*counterSeries
	gaugeSeries     map[string]*gaugeSeries
	histogramSeries map[string]*histogramSeries

	counterEvents   chan *counterEvent
	gaugeEvents     chan *gaugeEvent
	histogramEvents chan *histogramEvent

	mux      sync.RWMutex
	stopChan chan struct{}
	stopped  bool
	wg       sync.WaitGroup
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
		sender:          metricSender,
		interval:        interval,
		counterSeries:   make(map[string]*counterSeries),
		gaugeSeries:     make(map[string]*gaugeSeries),
		histogramSeries: make(map[string]*histogramSeries),
		counterEvents:   make(chan *counterEvent, 1000),
		gaugeEvents:     make(chan *gaugeEvent, 1000),
		histogramEvents: make(chan *histogramEvent, 1000),
		mux:             sync.RWMutex{},
		stopChan:        make(chan struct{}),
		stopped:         false,
		wg:              sync.WaitGroup{},
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
	c.stopped = true
	close(c.stopChan)
	c.wg.Wait()

	close(c.counterEvents)
	close(c.gaugeEvents)
	close(c.histogramEvents)

	c.processAfterShutdown()
	c.sendAfterShutdown()

	c.sender.stop()
}

// addCounter adds a counter event to the collector
func (c *metricCollector) addCounter(event *counterEvent) {
	fmt.Println("Adding counter event", event)
	if c.stopped {
		return
	}
	c.counterEvents <- event
}

// addGauge adds a gauge event to the collector
func (c *metricCollector) addGauge(event *gaugeEvent) {
	fmt.Println("Adding gauge event", event)
	if c.stopped {
		return
	}
	c.gaugeEvents <- event
}

// addHistogram adds a histogram event to the collector
func (c *metricCollector) addHistogram(event *histogramEvent) {
	fmt.Println("Adding histogram event", event)
	if c.stopped {
		return
	}
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

	now := time.Now()
	nextInterval := now.Truncate(c.interval).Add(c.interval)
	firstTriggerTime := nextInterval.Add(1 * time.Second)
	fmt.Println("First trigger time", firstTriggerTime)

	if firstTriggerTime.Before(now) {
		firstTriggerTime = nextInterval.Add(c.interval).Add(50 * time.Millisecond)
	}

	durationUntilFirstTrigger := firstTriggerTime.Sub(now)
	durationUntilFirstTrigger = time.Duration(max(0, int64(durationUntilFirstTrigger)))
	fmt.Println("Duration until first trigger", durationUntilFirstTrigger)

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
			fmt.Println("First tick time", firstTickTime)
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
					fmt.Println("Tick time", tickTime)
					intervalToProcess = tickTime.Truncate(c.interval).Add(-c.interval)
					fmt.Println("Interval to process", intervalToProcess)
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
func (c *metricCollector) processCounterEvent(event *counterEvent) {
	c.mux.Lock()
	defer c.mux.Unlock()

	identifier := newMetricIdentifier(event.name, event.tags)
	identifierString := identifier.String()

	if series, exists := c.counterSeries[identifierString]; exists {
		series.value += event.value
	} else {
		series := &counterSeries{
			name:  event.name,
			tags:  event.tags,
			value: event.value,
		}
		c.counterSeries[identifierString] = series
	}
}

// processGaugeEvent handles processing a single gauge event
func (c *metricCollector) processGaugeEvent(event *gaugeEvent) {
	c.mux.Lock()
	defer c.mux.Unlock()

	identifier := newMetricIdentifier(event.name, event.tags)
	identifierString := identifier.String()

	if series, exists := c.gaugeSeries[identifierString]; exists {
		switch event.mode {
		case GaugeModeInc:
			series.value += event.value
		case GaugeModeDec:
			series.value -= event.value
		case GaugeModeSet:
			series.value = event.value
		default:
			series.value = event.value
		}
	} else {
		series := &gaugeSeries{
			name:  event.name,
			tags:  event.tags,
			value: 0,
		}
		switch event.mode {
		case GaugeModeInc:
			series.value += event.value
		case GaugeModeDec:
			series.value -= event.value
		case GaugeModeSet:
			series.value = event.value
		default:
			series.value = event.value
		}
		c.gaugeSeries[identifierString] = series
	}
}

// processHistogramEvent handles processing a single histogram event
func (c *metricCollector) processHistogramEvent(event *histogramEvent) {
	c.mux.Lock()
	defer c.mux.Unlock()

	identifier := newMetricIdentifier(event.name, event.tags)
	identifierString := identifier.String()

	if series, exists := c.histogramSeries[identifierString]; exists {
		series.values = append(series.values, event.value)
	} else {
		series := &histogramSeries{
			name:   event.name,
			values: []float64{event.value},
			tags:   event.tags,
		}
		c.histogramSeries[identifierString] = series
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

	processedHistograms := 0
	for event := range c.histogramEvents {
		c.processHistogramEvent(event)
		processedHistograms++
	}
}

// sendMetricsForInterval sends the metrics for the interval
func (c *metricCollector) sendMetricsForInterval(intervalStart time.Time) {
	fmt.Println("Sending metrics for interval", intervalStart)
	c.mux.Lock()
	metricsToSend := c.aggregateMetrics(intervalStart)
	c.mux.Unlock()

	if metricsToSend != nil {
		c.sender.sendAggregatedMetrics(metricsToSend)
	}
}

// sendAfterShutdown sends all metrics currently held in buckets.
func (c *metricCollector) sendAfterShutdown() {
	c.mux.Lock()
	intervalStart := time.Now().Truncate(c.interval)
	metricsToSend := c.aggregateMetrics(intervalStart)
	c.resetMetrics()
	c.mux.Unlock()

	if metricsToSend != nil {
		c.sender.sendAggregatedMetrics(metricsToSend)
	}
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
	tags := make([]string, 0, len(m.tags))
	for k, v := range m.tags {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(tags)
	return strings.Join(append(parts, tags...), "_")
}

// aggregateMetrics creates a snapshot of the metrics for the given interval
func (c *metricCollector) aggregateMetrics(
	timestamp time.Time,
) *aggregatedMetrics {
	aggregatedMetrics := newAggregatedMetrics()

	for _, counter := range c.counterSeries {
		aggregatedMetrics.counterMetrics = append(aggregatedMetrics.counterMetrics, &counterMessage{
			Timestamp:  timestamp,
			MetricName: counter.name,
			Value:      counter.value,
			Tags:       counter.tags,
		})
	}

	for _, gauge := range c.gaugeSeries {
		aggregatedMetrics.gaugeMetrics = append(aggregatedMetrics.gaugeMetrics, &gaugeMessage{
			Timestamp:  timestamp,
			MetricName: gauge.name,
			Value:      gauge.value,
			Tags:       gauge.tags,
		})
	}

	for _, histogram := range c.histogramSeries {
		aggregatedMetrics.histogramMetrics = append(aggregatedMetrics.histogramMetrics, &histogramMessage{
			Timestamp:  timestamp,
			MetricName: histogram.name,
			Values:     histogram.values,
			Tags:       histogram.tags,
		})
	}

	return aggregatedMetrics
}

// resetMetrics resets the metrics for the given interval
func (c *metricCollector) resetMetrics() {
	c.mux.Lock()
	defer c.mux.Unlock()
	for _, counter := range c.counterSeries {
		counter.value = 0
	}
	for _, histogram := range c.histogramSeries {
		histogram.values = []float64{}
	}
}
