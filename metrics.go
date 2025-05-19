package vigilant

// Metric captures a metric
//
// Use this function when you want to capture a metric.
//
// Example:
//
//	Metric("my_metric", 1.0, vigilant.Tag("env", "prod"))
func Metric(name string, value float64, tags ...MetricTag) {
	if gateNilGlobalInstance() || value < 0 {
		return
	}

	metric := createMetricMessage(name, value, tags...)
	if metric == nil {
		return
	}

	globalInstance.captureMetric(metric)
}

// DEPRECATED: Use Metric instead
// MetricCounter captures a counter metric
func MetricCounter(name string, value float64, tags ...MetricTag) {
	if gateNilGlobalInstance() || value < 0 {
		return
	}

	counter := createCounterEvent(name, value, tags...)
	if counter == nil {
		return
	}

	globalInstance.captureCounter(counter)
}

// DEPRECATED: Use Metric instead
// MetricGauge captures a gauge metric
func MetricGauge(name string, value float64, mode GaugeMode, tags ...MetricTag) {
	if gateNilGlobalInstance() || value < 0 {
		return
	}

	gauge := createGaugeEvent(name, value, mode, tags...)
	if gauge == nil {
		return
	}

	globalInstance.captureGauge(gauge)
}

// DEPRECATED: Use Metric instead
// MetricHistogram captures a histogram metric
func MetricHistogram(name string, value float64, tags ...MetricTag) {
	if gateNilGlobalInstance() || value < 0 {
		return
	}

	histogram := createHistogramEvent(name, value, tags...)
	if histogram == nil {
		return
	}

	globalInstance.captureHistogram(histogram)
}
