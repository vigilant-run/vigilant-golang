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

// messageType represents the type of the message
type messageType string

const (
	messageTypeLog    messageType = "logs"
	messageTypeError  messageType = "errors"
	messageTypeMetric messageType = "metrics"
	messageTypeAlert  messageType = "alerts"
)

// messageBatch represents a batch of logs
type messageBatch struct {
	Token   string           `json:"token"`
	Type    messageType      `json:"type"`
	Logs    []*logMessage    `json:"logs,omitempty"`
	Errors  []*errorMessage  `json:"errors,omitempty"`
	Metrics []*metricMessage `json:"metrics,omitempty"`
	Alerts  []*alertMessage  `json:"alerts,omitempty"`
}

// logMessage represents a log message
type logMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Body       string            `json:"body"`
	Level      LogLevel          `json:"level"`
	Attributes map[string]string `json:"attributes"`
}

// errorMessage represents an error message
type errorMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Details    errorDetails      `json:"details"`
	Location   errorLocation     `json:"location"`
	Attributes map[string]string `json:"attributes"`
}

// errorLocation represents a location of an error
type errorLocation struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// errorDetails represents an error
type errorDetails struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Stacktrace string `json:"stacktrace"`
}

// alertMessage represents an alert message
type alertMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Title      string            `json:"title"`
	Attributes map[string]string `json:"attributes"`
}

// metricMessage represents a metric message
type metricMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Attributes map[string]string `json:"attributes"`
}
