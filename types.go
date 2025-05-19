package vigilant

import "time"

// LogLevel represents the severity of the log message
type LogLevel string

const (
	LEVEL_INFO  LogLevel = "INFO"
	LEVEL_WARN  LogLevel = "WARNING"
	LEVEL_ERROR LogLevel = "ERROR"
	LEVEL_DEBUG LogLevel = "DEBUG"
	LEVEL_TRACE LogLevel = "TRACE"
)

// GaugeMode is the mode of a gauge metric.
// It is used to specify the mode of a gauge metric.
type GaugeMode string

const (
	// GaugeModeSet sets the gauge to the given value.
	GaugeModeSet GaugeMode = "set"

	// GaugeModeInc increments the gauge by the given value.
	GaugeModeInc GaugeMode = "inc"

	// GaugeModeDec decrements the gauge by the given value.
	GaugeModeDec GaugeMode = "dec"
)

// messageBatch represents a batch of logs
type messageBatch struct {
	Token             string              `json:"token"`
	Logs              []*logMessage       `json:"logs,omitempty"`
	Metrics           []*metricMessage    `json:"metrics,omitempty"`
	MetricsCounters   []*counterMessage   `json:"metrics_counters,omitempty"`
	MetricsGauges     []*gaugeMessage     `json:"metrics_gauges,omitempty"`
	MetricsHistograms []*histogramMessage `json:"metrics_histograms,omitempty"`
}

// logMessage represents a log message
type logMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Body       string            `json:"body"`
	Level      LogLevel          `json:"level"`
	Attributes map[string]string `json:"attributes"`
}

// metricMessage represents a metric message
type metricMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Attributes map[string]string `json:"attributes"`
}

// counterMessage represents a counter metric message
type counterMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Tags       map[string]string `json:"tags"`
}

// gaugeMessage represents a gauge metric message
type gaugeMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Tags       map[string]string `json:"tags"`
}

// histogramMessage represents a histogram metric message
type histogramMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	MetricName string            `json:"metric_name"`
	Tags       map[string]string `json:"tags"`
	Values     []float64         `json:"values"`
}

// aggregatedMetrics represents a collection of counter and gauge metrics
type aggregatedMetrics struct {
	counterMetrics   []*counterMessage
	gaugeMetrics     []*gaugeMessage
	histogramMetrics []*histogramMessage
}

// newAggregatedMetrics creates a new aggregatedMetrics
func newAggregatedMetrics() *aggregatedMetrics {
	return &aggregatedMetrics{
		counterMetrics:   make([]*counterMessage, 0),
		gaugeMetrics:     make([]*gaugeMessage, 0),
		histogramMetrics: make([]*histogramMessage, 0),
	}
}

// internal counter event
type counterEvent struct {
	timestamp time.Time
	name      string
	value     float64
	tags      map[string]string
}

// internal gauge event
type gaugeEvent struct {
	timestamp time.Time
	name      string
	value     float64
	mode      GaugeMode
	tags      map[string]string
}

// internal histogram event
type histogramEvent struct {
	timestamp time.Time
	name      string
	value     float64
	tags      map[string]string
}

// counterSeries represents a series of counter metrics
type counterSeries struct {
	name  string
	tags  map[string]string
	value float64
}

// gaugeSeries represents a series of gauge metrics
type gaugeSeries struct {
	name  string
	tags  map[string]string
	value float64
}

// histogramSeries represents a series of histogram metrics
type histogramSeries struct {
	name   string
	tags   map[string]string
	values []float64
}
