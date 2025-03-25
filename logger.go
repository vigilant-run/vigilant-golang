package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LoggerConfig is the configuration for the logger
type LoggerConfig struct {
	Name        string
	Endpoint    string
	Token       string
	Passthrough bool
	Insecure    bool
	Noop        bool
}

// LoggerConfigBuilder is the builder for the logger configuration
type LoggerConfigBuilder struct {
	Name        string
	Endpoint    string
	Token       string
	Passthrough bool
	Insecure    bool
	Noop        bool
}

// NewLoggerConfigBuilder creates a new logger configuration builder
func NewLoggerConfigBuilder() *LoggerConfigBuilder {
	return &LoggerConfigBuilder{}
}

// WithName sets the name of the logger
func (b *LoggerConfigBuilder) WithName(name string) *LoggerConfigBuilder {
	b.Name = name
	return b
}

// WithEndpoint sets the endpoint of the logger
func (b *LoggerConfigBuilder) WithEndpoint(endpoint string) *LoggerConfigBuilder {
	b.Endpoint = endpoint
	return b
}

// WithToken sets the token of the logger
func (b *LoggerConfigBuilder) WithToken(token string) *LoggerConfigBuilder {
	b.Token = token
	return b
}

// WithPassthrough sets the passthrough flag of the logger
func (b *LoggerConfigBuilder) WithPassthrough() *LoggerConfigBuilder {
	b.Passthrough = true
	return b
}

// WithInsecure sets the insecure flag of the logger
func (b *LoggerConfigBuilder) WithInsecure() *LoggerConfigBuilder {
	b.Insecure = true
	return b
}

// WithNoop sets the noop flag of the logger
func (b *LoggerConfigBuilder) WithNoop() *LoggerConfigBuilder {
	b.Noop = true
	return b
}

// Build builds the logger configuration
func (b *LoggerConfigBuilder) Build() *LoggerConfig {
	config := &LoggerConfig{
		Name:        b.Name,
		Endpoint:    b.Endpoint,
		Token:       b.Token,
		Passthrough: b.Passthrough,
		Insecure:    b.Insecure,
		Noop:        b.Noop,
	}

	if b.Name == "" {
		config.Name = "service-name"
	}

	if b.Endpoint == "" {
		config.Endpoint = "ingress.vigilant.run"
	}

	if b.Token == "" {
		config.Token = "tk_1234567890"
	}

	return config
}

// InitLogger initializes the logger
func InitLogger(config *LoggerConfig) {
	globalLogger = NewLogger(config.Name, config.Endpoint, config.Token, config.Passthrough, config.Insecure, config.Noop)
}

// InitNoopLogger is convenience function for initializing a noop logger
func InitNoopLogger() {
	globalLogger = NewLogger("", "", "", false, false, true)
}

// ShutdownLogger shuts down the logger
func ShutdownLogger() error {
	return globalLogger.Shutdown()
}

// LogInfo logs a message at info level
func LogInfo(message string) {
	globalLogger.Info(message)
}

// LogWarn logs a message at warn level
func LogWarn(message string) {
	globalLogger.Warn(message)
}

// LogError logs a message at error level
func LogError(message string, err error) {
	globalLogger.Error(message, err)
}

// LogDebug logs a message at debug level
func LogDebug(message string) {
	globalLogger.Debug(message)
}

// LogInfoAttrs logs a message at info level with attributes
func LogInfoAttrs(message string, attrs map[string]string) {
	attrsList := make([]Attribute, 0, len(attrs))
	for k, v := range attrs {
		attrsList = append(attrsList, NewAttribute(k, v))
	}
	globalLogger.Info(message, attrsList...)
}

// LogWarnAttrs logs a message at warn level with attributes
func LogWarnAttrs(message string, attrs map[string]string) {
	attrsList := make([]Attribute, 0, len(attrs))
	for k, v := range attrs {
		attrsList = append(attrsList, NewAttribute(k, v))
	}
	globalLogger.Warn(message, attrsList...)
}

// LogErrorAttrs logs a message at error level with attributes
func LogErrorAttrs(message string, err error, attrs map[string]string) {
	attrsList := make([]Attribute, 0, len(attrs))
	for k, v := range attrs {
		attrsList = append(attrsList, NewAttribute(k, v))
	}
	globalLogger.Error(message, err, attrsList...)
}

// LogDebugAttrs logs a message at debug level with attributes
func LogDebugAttrs(message string, attrs map[string]string) {
	attrsList := make([]Attribute, 0, len(attrs))
	for k, v := range attrs {
		attrsList = append(attrsList, NewAttribute(k, v))
	}
	globalLogger.Debug(message, attrsList...)
}

var globalLogger *logger

// logger is the logger for the Vigilant platform
type logger struct {
	name        string
	endpoint    string
	token       string
	passthrough bool
	insecure    bool
	noop        bool

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
	noop bool,
) *logger {
	var formattedEndpoint string
	if insecure {
		formattedEndpoint = fmt.Sprintf("http://%s/api/message", endpoint)
	} else {
		formattedEndpoint = fmt.Sprintf("https://%s/api/message", endpoint)
	}

	logger := &logger{
		name:        name,
		endpoint:    formattedEndpoint,
		token:       token,
		passthrough: passthrough,
		insecure:    insecure,
		noop:        noop,
		logsQueue:   make(chan *logMessage, 1000),
		batchStop:   make(chan struct{}),
		wg:          sync.WaitGroup{},
	}

	logger.startBatcher()
	return logger
}

// Debug logs a message at DEBUG level
func (l *logger) Debug(message string, attrs ...Attribute) {
	l.log(LEVEL_DEBUG, message, nil, attrs...)
}

// Warn logs a message at WARN level
func (l *logger) Warn(message string, attrs ...Attribute) {
	l.log(LEVEL_WARN, message, nil, attrs...)
}

// Info logs a message at INFO level
func (l *logger) Info(message string, attrs ...Attribute) {
	l.log(LEVEL_INFO, message, nil, attrs...)
}

// Error logs a message at ERROR level
func (l *logger) Error(message string, err error, attrs ...Attribute) {
	l.log(LEVEL_ERROR, message, err, attrs...)
}

// Shutdown shuts down the logger
func (l *logger) Shutdown() error {
	l.stopBatcher()

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	<-done
	return nil
}

// log queues a log message to be sent to Vigilant
func (l *logger) log(level logLevel, message string, err error, attrs ...Attribute) {
	attrsMap := make(map[string]string)
	for _, attr := range attrs {
		attrsMap[attr.Key] = attr.Value
	}

	if err != nil {
		attrsMap["error"] = err.Error()
	}

	attrsMap["service.name"] = l.name

	l.logPassthrough(level, message, attrsMap)
	if l.noop {
		return
	}

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
func (l *logger) startBatcher() {
	l.wg.Add(1)
	go l.runBatcher()
}

// runBatcher is the batcher goroutine
func (l *logger) runBatcher() {
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
func (l *logger) stopBatcher() {
	close(l.batchStop)
}

// sendBatch sends a batch of logs
func (l *logger) sendBatch(logs []*logMessage) {
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

// logPassthrough logs a message to the console
func (l *logger) logPassthrough(level logLevel, message string, attrs map[string]string) {
	if !l.passthrough {
		return
	}

	fmt.Printf("[%s] %s %s\n", level, message, formatAttributes(attrs))
}

// formatAttributes formats the attributes
func formatAttributes(attrs map[string]string) string {
	attrsStr := ""
	i := 0
	for k, v := range attrs {
		if i > 0 {
			attrsStr += ", "
		}
		attrsStr += fmt.Sprintf("%s: %s", k, v)
		i++
	}
	return fmt.Sprintf("{%s}", attrsStr)
}
