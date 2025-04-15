package vigilant

// Counter captures a counter metric
func Counter(name string, value float64, tags map[string]string) {
	if gateNilAgent() {
		return
	}

	globalAgent.captureCounter(name, value, tags)
}

// Gauge captures a gauge metric
func Gauge(name string, value float64, tags map[string]string) {
	if gateNilAgent() {
		return
	}

	globalAgent.captureGauge(name, value, tags)
}

// Histogram captures a histogram metric
func Histogram(name string, value float64, tags map[string]string) {
	if gateNilAgent() {
		return
	}

	globalAgent.captureHistogram(name, value, tags)
}
