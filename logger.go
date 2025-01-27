package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LoggerBuilder are the options for the Logger
type LoggerBuilder struct {
	name        string
	endpoint    string
	token       string
	passthrough bool
	insecure    bool
}

// NewLoggerBuilder creates a new LoggerBuilder
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		name:        "go-server",
		endpoint:    "ingress.vigilant.run",
		token:       "tk_1234567890",
		passthrough: true,
		insecure:    false,
	}
}

// WithName adds the service name to the logger
func (o *LoggerBuilder) WithName(name string) *LoggerBuilder {
	o.name = name
	return o
}

// WithEndpoint adds the endpoint to the logger
func (o *LoggerBuilder) WithEndpoint(endpoint string) *LoggerBuilder {
	o.endpoint = endpoint
	return o
}

// WithToken adds the token to the logger
func (o *LoggerBuilder) WithToken(token string) *LoggerBuilder {
	o.token = token
	return o
}

// WithPassthrough also logs fmt.Println
func (o *LoggerBuilder) WithPassthrough() *LoggerBuilder {
	o.passthrough = true
	return o
}

// WithInsecure disables TLS verification
func (o *LoggerBuilder) WithInsecure() *LoggerBuilder {
	o.insecure = true
	return o
}

// Build builds the logger
func (o *LoggerBuilder) Build() *Logger {
	return NewLogger(o.name, o.endpoint, o.token, o.passthrough, o.insecure)
}

// Logger is the logger for the Vigilant platform
type Logger struct {
	name        string
	endpoint    string
	token       string
	passthrough bool
	insecure    bool

	logsQueue chan *logMessage
	batchStop chan struct{}
	wg        sync.WaitGroup
}

// NewLogger creates a new Logger
func NewLogger(
	name string,
	endpoint string,
	token string,
	passthrough bool,
	insecure bool,
) *Logger {
	var formattedEndpoint string
	if insecure {
		formattedEndpoint = fmt.Sprintf("http://%s/api/message", endpoint)
	} else {
		formattedEndpoint = fmt.Sprintf("https://%s/api/message", endpoint)
	}

	logger := &Logger{
		endpoint:    formattedEndpoint,
		token:       token,
		passthrough: passthrough,
		insecure:    insecure,
		logsQueue:   make(chan *logMessage, 1000),
		batchStop:   make(chan struct{}),
		wg:          sync.WaitGroup{},
	}

	logger.startBatcher()
	return logger
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(message string, attrs ...Attribute) {
	l.log(LEVEL_DEBUG, message, nil, attrs...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(message string, attrs ...Attribute) {
	l.log(LEVEL_WARN, message, nil, attrs...)
}

// Info logs a message at INFO level
func (l *Logger) Info(message string, attrs ...Attribute) {
	l.log(LEVEL_INFO, message, nil, attrs...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(message string, err error, attrs ...Attribute) {
	l.log(LEVEL_ERROR, message, err, attrs...)
}

// Shutdown shuts down the logger: flushes any remaining logs and stops the batcher goroutine
func (l *Logger) Shutdown() error {
	l.stopBatcher()

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	<-done
	return nil
}

// log queues a log message to be sent to the Vigilant platform
func (l *Logger) log(level logLevel, message string, err error, attrs ...Attribute) {
	attrsMap := make(map[string]string)
	for _, attr := range attrs {
		attrsMap[attr.Key] = attr.Value
	}

	if err != nil {
		attrsMap["error"] = err.Error()
	}

	attrsMap["service.name"] = l.name

	select {
	case l.logsQueue <- &logMessage{
		Timestamp:  time.Now(),
		Body:       message,
		Level:      level,
		Attributes: attrsMap,
	}:
	default:
	}
}

// startBatcher starts the batcher goroutine
func (l *Logger) startBatcher() {
	l.wg.Add(1)
	go l.runBatcher()
}

// runBatcher is the batcher goroutine
func (l *Logger) runBatcher() {
	defer l.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var batch []*logMessage

	for {
		select {
		case <-l.batchStop:
			if len(batch) > 0 {
				l.sendBatch(batch)
			}
			return

		case msg := <-l.logsQueue:
			if msg == nil {
				continue
			}

			batch = append(batch, msg)
			if len(batch) >= maxBatchSize {
				l.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				l.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// stopBatcher closes the batchStop channel
func (l *Logger) stopBatcher() {
	close(l.batchStop)
}

// sendBatch sends a batch of logs to the Vigilant platform
func (l *Logger) sendBatch(logs []*logMessage) {
	if len(logs) == 0 {
		return
	}

	batch := &messageBatch{
		Token: l.token,
		Type:  messageTypeLog,
		Logs:  logs,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", l.endpoint, bytes.NewBuffer(batchBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// logLevel represents the severity of the log message
type logLevel string

const (
	LEVEL_INFO  logLevel = "INFO"
	LEVEL_WARN  logLevel = "WARNING"
	LEVEL_ERROR logLevel = "ERROR"
	LEVEL_DEBUG logLevel = "DEBUG"
)

// messageType represents the type of the message
type messageType string

const (
	messageTypeLog messageType = "logs"
)

// messageBatch represents a batch of logs
type messageBatch struct {
	Token string        `json:"token"`
	Type  messageType   `json:"type"`
	Logs  []*logMessage `json:"logs,omitempty"`
}

// logMessage represents a log message
type logMessage struct {
	Timestamp  time.Time         `json:"timestamp"`
	Body       string            `json:"body"`
	Level      logLevel          `json:"level"`
	Attributes map[string]string `json:"attributes"`
}
