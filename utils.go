package vigilant

import (
	"bytes"
	"fmt"
	"time"
)

// keyValsToMap formats a list of key-value pairs into a map
// it is a utility function for some of the observability functions
func keyValsToMap(keyVals ...any) (map[string]string, error) {
	attrs := make(map[string]string)
	if len(keyVals)%2 != 0 {
		return nil, fmt.Errorf("invalid number of key-value pairs")
	}
	for i := 0; i < len(keyVals); i += 2 {
		key := fmt.Sprintf("%v", keyVals[i])
		value := fmt.Sprintf("%v", keyVals[i+1])
		attrs[key] = value
	}
	return attrs, nil
}

// attributesToMap formats a list of attributes into a map
// it is a utility function for some of the observability functions
func attributesToMap(attributes ...Attribute) (map[string]string, error) {
	attrs := make(map[string]string)
	for _, attribute := range attributes {
		attrs[attribute.Key] = attribute.Value
	}
	return attrs, nil
}

// prettyPrintAttributes pretty prints a map of attributes
func prettyPrintAttributes(attrs map[string]string) string {
	var sb bytes.Buffer
	for key, value := range attrs {
		sb.WriteString(fmt.Sprintf("%s=%s ", key, value))
	}
	return sb.String()
}

// getEndpoint returns the endpoint for the given config
func getEndpoint(config *VigilantConfig) string {
	var prefix string
	if config.Insecure {
		prefix = "http://"
	} else {
		prefix = "https://"
	}
	return prefix + config.Endpoint
}

// isLevelEnabled checks if the given level is enabled
func isLevelEnabled(level LogLevel, minLevel LogLevel) bool {
	levelInt := getLevelInt(level)
	minLevelInt := getLevelInt(minLevel)
	return levelInt >= minLevelInt
}

// getLevelInt returns the integer value of the given level
func getLevelInt(level LogLevel) int {
	switch level {
	case LEVEL_ERROR:
		return 5
	case LEVEL_WARN:
		return 4
	case LEVEL_INFO:
		return 3
	case LEVEL_DEBUG:
		return 2
	case LEVEL_TRACE:
		return 1
	default:
		return 0
	}
}

// globalEmitNilError is a flag to emit the nil error only once
var globalEmitNilError = true

// gateNilGlobalInstance checks if the Vigilant instance is nil
func gateNilGlobalInstance() bool {
	if globalInstance != nil {
		return false
	}
	if globalEmitNilError {
		fmt.Printf("\n[ERROR] Vigilant is not initialized.\n\tPlease call vigilant.Init() before using Vigilant.\n\tDocs: https://docs.vigilant.run/\n")
		globalEmitNilError = false
	}
	return true
}

// createLogMessage creates a log message from the given parameters
func createLogMessage(level LogLevel, message string, attributes map[string]string) *logMessage {
	deduplicatedAttributes := deduplicateAttributes(attributes)
	return &logMessage{
		Timestamp:  time.Now(),
		Level:      level,
		Body:       message,
		Attributes: deduplicatedAttributes,
	}
}

// createCounterEvent creates a counter event from the given parameters
func createCounterEvent(name string, value float64, tags ...MetricTag) *counterEvent {
	deduplicatedTags := deduplicateTags(tags)
	return &counterEvent{
		timestamp: time.Now(),
		name:      name,
		value:     value,
		tags:      deduplicatedTags,
	}
}

// createGaugeEvent creates a gauge event from the given parameters
func createGaugeEvent(name string, value float64, mode GaugeMode, tags ...MetricTag) *gaugeEvent {
	deduplicatedTags := deduplicateTags(tags)
	return &gaugeEvent{
		timestamp: time.Now(),
		name:      name,
		value:     value,
		mode:      mode,
		tags:      deduplicatedTags,
	}
}

// createHistogramEvent creates a histogram event from the given parameters
func createHistogramEvent(name string, value float64, tags ...MetricTag) *histogramEvent {
	deduplicatedTags := deduplicateTags(tags)
	return &histogramEvent{
		timestamp: time.Now(),
		name:      name,
		value:     value,
		tags:      deduplicatedTags,
	}
}

// deduplicateAttributes deduplicates the attributes
func deduplicateAttributes(attributes map[string]string) map[string]string {
	deduplicated := make(map[string]string)
	for key, value := range attributes {
		if _, ok := deduplicated[key]; !ok {
			deduplicated[key] = value
		}
	}
	return deduplicated
}

// deduplicateTags deduplicates the tags
func deduplicateTags(tags []MetricTag) map[string]string {
	deduplicated := make(map[string]string)
	for _, tag := range tags {
		if _, ok := deduplicated[tag.Key]; !ok {
			deduplicated[tag.Key] = tag.Value
		}
	}
	return deduplicated
}
