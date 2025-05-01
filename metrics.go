package vigilant

import "fmt"

// Counter captures a counter metric
func MetricCounter(name string, value float64, tags ...MetricTag) {
	if gateNilAgent() {
		return
	}

	tagsMap, err := tagsToMap(tags...)
	if err != nil {
		fmt.Printf("error formatting tags: %v\n", err)
		return
	}

	globalAgent.captureCounter(name, value, tagsMap)
}

// Gauge captures a gauge metric
func MetricGauge(name string, value float64, tags ...MetricTag) {
	if gateNilAgent() {
		return
	}

	tagsMap, err := tagsToMap(tags...)
	if err != nil {
		fmt.Printf("error formatting tags: %v\n", err)
		return
	}

	globalAgent.captureGauge(name, value, tagsMap)
}

// Histogram captures a histogram metric
func MetricHistogram(name string, value float64, tags ...MetricTag) {
	if gateNilAgent() {
		return
	}

	tagsMap, err := tagsToMap(tags...)
	if err != nil {
		fmt.Printf("error formatting tags: %v\n", err)
		return
	}

	globalAgent.captureHistogram(name, value, tagsMap)
}
