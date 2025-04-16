package vigilant

import (
	"bytes"
	"fmt"
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

// gateNilAgent checks if the agent is nil
func gateNilAgent() bool {
	if globalAgent == nil {
		fmt.Printf("\n[ERROR] The Vigilant agent is not initialized.\n\tPlease call vigilant.Init() before using the agent.\n\tDocs: https://docs.vigilant.run/overview\n")
		return true
	}
	return false
}
