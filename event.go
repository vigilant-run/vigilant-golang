package vigilant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

const EVENTS_PATH = "/api/events"

// EventHandlerOptions are the options for the EventHandler
type EventHandlerOptions struct {
	url      string
	token    string
	insecure bool
	name     string
	noop     bool
}

// NewEventHandlerOptions creates a new EventHandlerOptions
func NewEventHandlerOptions(opts ...EventHandlerOption) *EventHandlerOptions {
	options := &EventHandlerOptions{
		url:      "https://errors.vigilant.run" + EVENTS_PATH,
		token:    "tk_1234567890",
		name:     "go-server",
		noop:     false,
		insecure: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// EventHandlerOption is a function that configures the EventHandlerOptions
type EventHandlerOption func(*EventHandlerOptions)

// WithErrorHandlerName sets the name of the service
func WithEventHandlerName(name string) EventHandlerOption {
	return func(opts *EventHandlerOptions) {
		opts.name = name
	}
}

// WithErrorHandlerURL sets the server URL for the error handler
func WithEventHandlerURL(url string) EventHandlerOption {
	return func(opts *EventHandlerOptions) {
		opts.url = url + EVENTS_PATH
	}
}

// WithErrorHandlerToken sets the token for authentication
func WithEventHandlerToken(token string) EventHandlerOption {
	return func(opts *EventHandlerOptions) {
		opts.token = token
	}
}

// WithErrorHandlerInsecure disables TLS verification
func WithEventHandlerInsecure() EventHandlerOption {
	return func(opts *EventHandlerOptions) {
		opts.insecure = true
	}
}

// WithErrorHandlerNoop disables the error handler
func WithEventHandlerNoop() EventHandlerOption {
	return func(opts *EventHandlerOptions) {
		opts.noop = true
	}
}

// EventHandler captures and sends events to the event server
type EventHandler struct {
	client *http.Client

	options *EventHandlerOptions

	newEvents     chan *internalEvent
	batchedEvents []*internalEvent
	stop          chan struct{}
	mux           sync.Mutex
	wg            sync.WaitGroup
}

// NewErrorHandler creates a new ErrorHandler
func NewEventHandler(opts ...EventHandlerOption) (*EventHandler, error) {
	options := &EventHandlerOptions{
		url:  "https://errors.vigilant.run" + EVENTS_PATH,
		name: "go-server",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.url == "" {
		return nil, fmt.Errorf("error handler URL is empty")
	}

	if options.token == "" {
		return nil, fmt.Errorf("error handler token is empty")
	}

	handler := &EventHandler{
		client:        &http.Client{Timeout: 5 * time.Second},
		options:       options,
		mux:           sync.Mutex{},
		stop:          make(chan struct{}),
		newEvents:     make(chan *internalEvent, 1000),
		batchedEvents: make([]*internalEvent, 0, 1000),
	}

	handler.start()

	return handler, nil
}

// CaptureMessage sends a message event to the event server
func (h *EventHandler) CaptureMessage(message string) error {
	if h.options.noop {
		return nil
	}

	select {
	case h.newEvents <- h.parseMessage(message):
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// CaptureError sends an error event to the event server
func (h *EventHandler) CaptureError(err error) error {
	if h.options.noop {
		return nil
	}

	select {
	case h.newEvents <- h.parseError(err):
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// Shutdown stops the error handler
func (h *EventHandler) Shutdown() {
	close(h.stop)
	h.wg.Wait()
}

// start starts the error handler
func (h *EventHandler) start() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-h.stop:
				h.processRemainingEvents()
				return
			case data := <-h.newEvents:
				h.mux.Lock()
				h.batchedEvents = append(h.batchedEvents, data)
				h.mux.Unlock()
			case <-ticker.C:
				h.mux.Lock()
				if len(h.batchedEvents) > 0 {
					err := h.sendBatch(context.Background())
					if err != nil {
						fmt.Printf("error sending batch: %v\n", err)
					}
				}
				h.mux.Unlock()
			}
		}
	}()
}

// processRemainingEvents handles any remaining events during shutdown
func (h *EventHandler) processRemainingEvents() {
	for {
		select {
		case data := <-h.newEvents:
			h.mux.Lock()
			h.batchedEvents = append(h.batchedEvents, data)
			h.mux.Unlock()
		default:
			h.mux.Lock()
			if len(h.batchedEvents) > 0 {
				err := h.sendBatch(context.Background())
				if err != nil {
					fmt.Printf("error sending final batch: %v\n", err)
				}
			}
			h.mux.Unlock()
			return
		}
	}
}

// sendBatch sends a batch of errors to the error server
func (h *EventHandler) sendBatch(ctx context.Context) error {
	if len(h.batchedEvents) == 0 {
		return nil
	}

	data, err := json.Marshal(h.batchedEvents)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.options.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-vigilant-token", h.options.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status code %d", resp.StatusCode)
	}

	h.batchedEvents = h.batchedEvents[:0]

	return nil
}

// parseMessage parses the message and returns the internal message structure
func (h *EventHandler) parseMessage(message string) *internalEvent {
	return &internalEvent{
		Timestamp:  time.Now().UTC(),
		Message:    &message,
		Exceptions: []exception{},
		Metadata:   h.getMetadata(),
	}
}

// parseError parses the error and returns the internal error structure
func (h *EventHandler) parseError(err error) *internalEvent {
	return &internalEvent{
		Timestamp:  time.Now().UTC(),
		Message:    nil,
		Exceptions: []exception{getException(err)},
		Metadata:   h.getMetadata(),
	}
}

// getMetadata returns the metadata for the given message
func (h *EventHandler) getMetadata() map[string]string {
	filename := getFilename(4)
	line := getFileline(4)
	function := getFunctionName(4)
	os := getOS()
	stackTrace := h.getStackTrace()
	arch := getArch()
	goVersion := getGoVersion()
	return map[string]string{
		"service":    h.options.name,
		"function":   function,
		"filename":   filename,
		"line":       strconv.Itoa(line),
		"os":         os,
		"arch":       arch,
		"go.version": goVersion,
		"stack":      stackTrace,
	}
}

// getStackTrace returns the stack trace for the given error
func (h *EventHandler) getStackTrace() string {
	return string(debug.Stack())
}

// getFilename returns the filename where the error occurred
func getFilename(skip int) string {
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return file
}

// getFunctionName returns the name of the function that called the given error
func getFunctionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return runtime.FuncForPC(pc).Name()
}

// getFileline returns the line number where the error occurred
func getFileline(skip int) int {
	_, _, line, ok := runtime.Caller(skip)
	if !ok {
		return 0
	}
	return line
}

// getOS returns the operating system
func getOS() string {
	return runtime.GOOS
}

// getArch returns the architecture
func getArch() string {
	return runtime.GOARCH
}

// getGoVersion returns the Go version
func getGoVersion() string {
	return runtime.Version()
}

// getException returns the exception for the given error
func getException(err error) exception {
	exception := exception{
		Type:  reflect.TypeOf(err).String(),
		Value: err.Error(),
		Stack: getStackFrames(),
	}
	return exception
}

// prohibitedModules are the modules that are not allowed to be sent to the event server
var prohibitedModules = []string{
	"runtime",
	"testing",
	"vendor",
	"third_party",
	"github.com/vigilant-run/vigilant-go",
}

// getStackFrames extracts stack frames from the current goroutine
func getStackFrames() []frame {
	pointers := make([]uintptr, 50)
	n := runtime.Callers(2, pointers)
	if n == 0 {
		return nil
	}

	frames := make([]frame, 0, n)
	callersFrames := runtime.CallersFrames(pointers[:n])

	for {
		callerFrame, more := callersFrames.Next()
		if !more {
			break
		}

		module, function := splitFunctionName(callerFrame.Function)

		isInternal := true
		for _, prohibitedModule := range prohibitedModules {
			if strings.Contains(module, prohibitedModule) {
				isInternal = false
				break
			}
		}

		frames = append(frames, frame{
			Function: function,
			Module:   module,
			File:     callerFrame.File,
			Line:     callerFrame.Line,
			Internal: isInternal,
		})

		if !more {
			break
		}
	}

	reversed := make([]frame, len(frames))
	for i := range frames {
		reversed[len(frames)-i-1] = frames[i]
	}

	return reversed
}

// splitFunctionName splits a package path-qualified function name
func splitFunctionName(name string) (string, string) {
	slash := strings.LastIndex(name, "/")
	if slash < 0 {
		slash = 0
	}

	dot := strings.LastIndex(name[slash:], ".")
	if dot < 0 {
		return "", name
	}

	return name[:slash+dot], name[slash+dot+1:]
}
