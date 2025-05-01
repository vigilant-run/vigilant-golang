package vigilant

// MetricCounter captures a counter metric
func MetricCounter(name string, value float64, tags ...MetricTag) {
	if gateNilAgent() || value < 0 {
		return
	}

	counter := createCounterEvent(name, value, tags...)
	if counter == nil {
		return
	}

	globalAgent.captureCounter(counter)
}

// MetricGauge captures a gauge metric
func MetricGauge(name string, value float64, mode GaugeMode, tags ...MetricTag) {
	if gateNilAgent() || value < 0 {
		return
	}

	gauge := createGaugeEvent(name, value, mode, tags...)
	if gauge == nil {
		return
	}

	globalAgent.captureGauge(gauge)
}

// MetricHistogram captures a histogram metric
func MetricHistogram(name string, value float64, tags ...MetricTag) {
	if gateNilAgent() || value < 0 {
		return
	}

	histogram := createHistogramEvent(name, value, tags...)
	if histogram == nil {
		return
	}

	globalAgent.captureHistogram(histogram)
}
