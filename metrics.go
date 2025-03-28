package vigilant

import "fmt"

// Metrics capture functions are used to capture metrics, they are viewable in the Vigilant Dashboard.

// ----------------------- //
// --- Metric Capture --- //
// ----------------------- //

// EmitMetric captures a metric and sends it to the agent
// Example:
// EmitMetric("my_metric", 1.0)
func EmitMetric(name string, value float64) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendMetric(name, value, nil)
}

// EmitMetricw captures a metric and sends it to the agent with attributes
// Example:
// EmitMetricw("my_metric", 1.0, "key1", "value1", "key2", "value2")
func EmitMetricw(name string, value float64, keyVals ...any) {
	if globalAgent == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendMetric(name, value, attrs)
}

// EmitMetrict captures a metric and sends it to the agent with typed attributes
// Example:
// EmitMetrict("my_metric", vigilant.Float64("value", 1.2345))
func EmitMetrict(name string, value float64, fields ...Field) {
	if globalAgent == nil {
		return
	}

	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendMetric(name, value, attrs)
}

// writeMetricPassthrough writes a metric to the agent
// this is an internal function that is used to write metrics to stdout
func writeMetricPassthrough(name string, value float64, attrs map[string]string) {
	if len(attrs) > 0 {
		fmt.Printf("[METRIC] name=%s value=%f %s\n", name, value, prettyPrintAttributes(attrs))
	} else {
		fmt.Printf("[METRIC] name=%s value=%f\n", name, value)
	}
}
