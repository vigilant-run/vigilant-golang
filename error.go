package vigilant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const ERRORS_PATH = "/api/errors"

// ErrorHandlerOptions are the options for the ErrorCaptureHandler
type ErrorHandlerOptions struct {
	url      string
	token    string
	insecure bool
	name     string
}

// ErrorHandlerOption is a function that configures the ErrorHandlerOptions
type ErrorHandlerOption func(*ErrorHandlerOptions)

// WithErrorHandlerName sets the name of the service
func WithErrorHandlerName(name string) ErrorHandlerOption {
	return func(opts *ErrorHandlerOptions) {
		opts.name = name
	}
}

// WithErrorHandlerURL sets the server URL for the error handler
func WithErrorHandlerURL(url string) ErrorHandlerOption {
	return func(opts *ErrorHandlerOptions) {
		opts.url = url
	}
}

// WithErrorHandlerToken sets the token for authentication
func WithErrorHandlerToken(token string) ErrorHandlerOption {
	return func(opts *ErrorHandlerOptions) {
		opts.token = token
	}
}

// WithErrorHandlerInsecure disables TLS verification
func WithErrorHandlerInsecure() ErrorHandlerOption {
	return func(opts *ErrorHandlerOptions) {
		opts.insecure = true
	}
}

// internalError is an internal error that is used to wrap errors
type internalError struct {
	Timestamp  time.Time   `json:"timestamp"`
	Error      string      `json:"error"`
	Attributes []Attribute `json:"attributes"`
}

// ErrorHandler captures and sends errors to the error server
type ErrorHandler struct {
	client *http.Client

	options *ErrorHandlerOptions

	newErrors     chan *internalError
	batchedErrors []*internalError
	stop          chan struct{}
	mux           sync.Mutex
	wg            sync.WaitGroup
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler(opts ...ErrorHandlerOption) (*ErrorHandler, error) {
	options := &ErrorHandlerOptions{
		url:  "https://errors.vigilant.run" + ERRORS_PATH,
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

	handler := &ErrorHandler{
		client:        &http.Client{Timeout: 5 * time.Second},
		options:       options,
		mux:           sync.Mutex{},
		stop:          make(chan struct{}),
		newErrors:     make(chan *internalError, 1000),
		batchedErrors: make([]*internalError, 0, 1000),
	}

	handler.start()

	return handler, nil
}

// Capture sends an error event to the error server
func (h *ErrorHandler) Capture(ctx context.Context, err error, attrs ...Attribute) error {
	select {
	case h.newErrors <- h.parseError(err, attrs...):
		return nil
	default:
		return fmt.Errorf("error channel is full")
	}
}

// Shutdown stops the error handler
func (h *ErrorHandler) Shutdown() {
	close(h.stop)
	h.wg.Wait()
}

// start starts the error handler
func (h *ErrorHandler) start() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-h.stop:
				h.processRemainingErrors()
				return
			case data := <-h.newErrors:
				h.mux.Lock()
				h.batchedErrors = append(h.batchedErrors, data)
				h.mux.Unlock()
			case <-ticker.C:
				h.mux.Lock()
				if len(h.batchedErrors) > 0 {
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

// processRemainingErrors handles any remaining errors during shutdown
func (h *ErrorHandler) processRemainingErrors() {
	for {
		select {
		case data := <-h.newErrors:
			h.mux.Lock()
			h.batchedErrors = append(h.batchedErrors, data)
			h.mux.Unlock()
		default:
			h.mux.Lock()
			if len(h.batchedErrors) > 0 {
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
func (h *ErrorHandler) sendBatch(ctx context.Context) error {
	if len(h.batchedErrors) == 0 {
		return nil
	}

	data, err := json.Marshal(h.batchedErrors)
	if err != nil {
		return fmt.Errorf("failed to marshal error payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.options.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-vigilant-token", h.options.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send error event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status code %d", resp.StatusCode)
	}

	h.batchedErrors = h.batchedErrors[:0]

	return nil
}

// parseError parses the error and returns the internal error structure
func (h *ErrorHandler) parseError(err error, attrs ...Attribute) *internalError {
	return &internalError{
		Timestamp:  time.Now().UTC(),
		Error:      err.Error(),
		Attributes: h.getErrorAttributes(err, attrs...),
	}
}

// getErrorAttributes returns the attributes for the given error
func (h *ErrorHandler) getErrorAttributes(err error, attrs ...Attribute) []Attribute {
	filename := getFilename(4)
	line := getFileline(4)
	function := getFunctionName(4)
	stack := h.getStackTrace(err)
	os := getOS()
	arch := getArch()
	goVersion := getGoVersion()
	allAttrs := []Attribute{
		NewAttribute("service", h.options.name),
		NewAttribute("function", function),
		NewAttribute("filename", filename),
		NewAttribute("line", line),
		NewAttribute("stack", stack),
		NewAttribute("os", os),
		NewAttribute("arch", arch),
		NewAttribute("go.version", goVersion),
	}
	return append(allAttrs, attrs...)
}

// getStackTrace returns the stack trace for the given error
func (h *ErrorHandler) getStackTrace(err error) string {
	if err == nil {
		return ""
	}
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
