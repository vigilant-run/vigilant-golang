package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorHandlerConfig is the configuration for the error handler
type ErrorHandlerConfig struct {
	Name        string
	Endpoint    string
	Token       string
	Passthrough bool
	Insecure    bool
	Noop        bool
}

// ErrorHandlerConfigBuilder is the builder for the error handler configuration
type ErrorHandlerConfigBuilder struct {
	Name        string
	Endpoint    string
	Token       string
	Passthrough bool
	Insecure    bool
	Noop        bool
}

// NewErrorHandlerConfigBuilder creates a new error handler configuration builder
func NewErrorHandlerConfigBuilder() *ErrorHandlerConfigBuilder {
	return &ErrorHandlerConfigBuilder{}
}

// WithName sets the name of the error handler
func (b *ErrorHandlerConfigBuilder) WithName(name string) *ErrorHandlerConfigBuilder {
	b.Name = name
	return b
}

// WithEndpoint sets the endpoint of the error handler
func (b *ErrorHandlerConfigBuilder) WithEndpoint(endpoint string) *ErrorHandlerConfigBuilder {
	b.Endpoint = endpoint
	return b
}

// WithToken sets the token of the error handler
func (b *ErrorHandlerConfigBuilder) WithToken(token string) *ErrorHandlerConfigBuilder {
	b.Token = token
	return b
}

// WithPassthrough sets the passthrough of the error handler
func (b *ErrorHandlerConfigBuilder) WithPassthrough() *ErrorHandlerConfigBuilder {
	b.Passthrough = true
	return b
}

// WithInsecure sets the insecure of the error handler
func (b *ErrorHandlerConfigBuilder) WithInsecure() *ErrorHandlerConfigBuilder {
	b.Insecure = true
	return b
}

// WithNoop sets the noop of the error handler
func (b *ErrorHandlerConfigBuilder) WithNoop() *ErrorHandlerConfigBuilder {
	b.Noop = true
	return b
}

// Build builds the error handler configuration
func (b *ErrorHandlerConfigBuilder) Build() *ErrorHandlerConfig {
	config := &ErrorHandlerConfig{
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

// InitErrorHandler initializes the error handler
func InitErrorHandler(config *ErrorHandlerConfig) {
	globalErrorHandler = newErrorHandler(config.Name, config.Endpoint, config.Token, config.Passthrough, config.Insecure, config.Noop)
}

// ShutdownErrorHandler shuts down the error handler
func ShutdownErrorHandler() error {
	return globalErrorHandler.shutdown()
}

// CaptureError captures an error
func CaptureError(err error, attrs ...Attribute) {
	if globalErrorHandler == nil {
		return
	}
	globalErrorHandler.capture(err, attrs...)
}

var globalErrorHandler *errorHandler

// errorHandler is a handler for errors
type errorHandler struct {
	name        string
	endpoint    string
	token       string
	passthrough bool
	insecure    bool
	noop        bool

	errorsQueue chan *errorMessage
	batchStop   chan struct{}
	wg          sync.WaitGroup
}

// newErrorHandler creates a new errorHandler
func newErrorHandler(
	name string,
	endpoint string,
	token string,
	passthrough bool,
	insecure bool,
	noop bool,
) *errorHandler {
	var formattedEndpoint string
	if insecure {
		formattedEndpoint = fmt.Sprintf("http://%s/api/message", endpoint)
	} else {
		formattedEndpoint = fmt.Sprintf("https://%s/api/message", endpoint)
	}

	errorHandler := &errorHandler{
		name:        name,
		endpoint:    formattedEndpoint,
		token:       token,
		passthrough: passthrough,
		insecure:    insecure,
		noop:        noop,
		errorsQueue: make(chan *errorMessage, 1000),
		batchStop:   make(chan struct{}),
	}

	errorHandler.startBatcher()
	return errorHandler
}

// capture captures an error and sends it to Vigilant
func (e *errorHandler) capture(err error, attrs ...Attribute) {
	if e.noop || err == nil {
		return
	}

	attrsMap := make(map[string]string)
	for _, attr := range attrs {
		attrsMap[attr.Key] = attr.Value
	}
	attrsMap["service.name"] = e.name

	select {
	case e.errorsQueue <- &errorMessage{
		Timestamp:  time.Now(),
		Details:    getDetails(err),
		Location:   getLocation(3),
		Attributes: attrsMap,
	}:
	default:
	}
}

// Shutdown shuts down the error handler
func (e *errorHandler) shutdown() error {
	e.stopBatcher()

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	<-done
	return nil
}

// startBatcher starts the batcher goroutine
func (e *errorHandler) startBatcher() {
	e.wg.Add(1)
	go e.runBatcher()
}

// runBatcher is the batcher goroutine
func (e *errorHandler) runBatcher() {
	defer e.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var batch []*errorMessage

	for {
		select {
		case <-e.batchStop:
			if len(batch) > 0 {
				e.sendBatch(batch)
			}
			return

		case msg := <-e.errorsQueue:
			if msg == nil {
				continue
			}

			batch = append(batch, msg)
			if len(batch) >= maxBatchSize {
				e.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				e.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// stopBatcher closes the batchStop channel
func (e *errorHandler) stopBatcher() {
	close(e.batchStop)
}

// sendBatch sends a batch of errors
func (e *errorHandler) sendBatch(errors []*errorMessage) {
	if len(errors) == 0 {
		return
	}

	batch := &messageBatch{
		Token:  e.token,
		Type:   messageTypeError,
		Errors: errors,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", e.endpoint, bytes.NewBuffer(batchBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// getDetails returns the details of an error
func getDetails(err error) errorDetails {
	stacktrace := buildStackTrace(5, err)
	return errorDetails{
		Type:       fmt.Sprintf("%T", err),
		Message:    err.Error(),
		Stacktrace: stacktrace,
	}
}

// getLocation returns the location of an error
func getLocation(skip int) errorLocation {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok || pc == 0 {
		return errorLocation{
			Function: "unknown",
			File:     "unknown",
			Line:     0,
		}
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return errorLocation{
			Function: "unknown",
			File:     file,
			Line:     line,
		}
	}

	fullName := getFunctionName(fn)
	if fullName == "" {
		fullName = "unknown"
	}

	return errorLocation{
		Function: fullName,
		File:     file,
		Line:     line,
	}
}

// getFunctionName returns the function name from a function
func getFunctionName(fn *runtime.Func) string {
	if fn == nil {
		return ""
	}

	fullName := fn.Name()
	if idx := strings.LastIndex(fullName, "/"); idx >= 0 {
		fullName = fullName[idx+1:]
	}

	if idx := strings.Index(fullName, "."); idx >= 0 {
		fullName = fullName[idx+1:]
	}

	return fullName
}

// buildStackTrace is a helper to gather the complete stack from the caller
func buildStackTrace(skip int, err error) string {
	pc := make([]uintptr, 32)
	n := runtime.Callers(skip, pc)
	pc = pc[:n]

	frames := runtime.CallersFrames(pc)
	var sb bytes.Buffer

	sb.WriteString(fmt.Sprintf("%T: %s\n", err, err.Error()))

	for {
		frame, more := frames.Next()
		funcName := frame.Function
		if funcName == "" {
			funcName = "unknown"
		}

		sb.WriteString(fmt.Sprintf("  File \"%s\", line %d, in %s\n", frame.File, frame.Line, funcName))
		if !more {
			break
		}
	}

	return sb.String()
}
