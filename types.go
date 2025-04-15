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

// messageBatch represents a batch of logs
type messageBatch struct {
	Token string        `json:"token"`
	Logs  []*logMessage `json:"logs,omitempty"`
}

// logMessage represents a log message
type logMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Body       string            `json:"body"`
	Level      LogLevel          `json:"level"`
	Attributes map[string]string `json:"attributes"`
}
