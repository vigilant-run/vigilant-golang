package vigilant

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type customError struct {
	msg   string
	stack []uintptr
}

func (e *customError) Error() string {
	return e.msg
}

func newCustomError(msg string) *customError {
	return &customError{
		msg:   msg,
		stack: make([]uintptr, 1),
	}
}

func TestNewErrorHandler(t *testing.T) {
	tests := []struct {
		name    string
		opts    []ErrorHandlerOption
		wantErr bool
	}{
		{
			name:    "no options",
			opts:    []ErrorHandlerOption{},
			wantErr: true,
		},
		{
			name: "with valid options",
			opts: []ErrorHandlerOption{
				WithErrorHandlerURL("https://test.com"),
				WithErrorHandlerToken("test-token"),
				WithErrorHandlerName("test-service"),
			},
			wantErr: false,
		},
		{
			name: "missing token",
			opts: []ErrorHandlerOption{
				WithErrorHandlerURL("https://test.com"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewErrorHandler(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewErrorHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				handler.Shutdown()
			}
		})
	}
}

func TestErrorHandlerCapture(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type header to be application/json")
		}
		if r.Header.Get("X-Vigilant-Token") != "test-token" {
			t.Error("Expected X-Vigilant-Token header to be test-token")
		}

		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewErrorHandler(
		WithErrorHandlerURL(server.URL),
		WithErrorHandlerToken("test-token"),
		WithErrorHandlerName("test-service"),
	)
	if err != nil {
		t.Fatalf("Failed to create ErrorHandler: %v", err)
	}

	defer handler.Shutdown()

	tests := []struct {
		name    string
		err     error
		attrs   []Attribute
		wantErr bool
	}{
		{
			name:    "basic error",
			err:     fmt.Errorf("test error"),
			attrs:   []Attribute{{Key: "test", Value: true}},
			wantErr: false,
		},
		{
			name:    "custom error with stack",
			err:     newCustomError("custom error"),
			attrs:   []Attribute{{Key: "test", Value: true}},
			wantErr: false,
		},
		{
			name:    "nil metadata",
			err:     fmt.Errorf("error without metadata"),
			attrs:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captureErr := handler.Capture(context.Background(), tt.err, tt.attrs...)
			if (captureErr != nil) != tt.wantErr {
				t.Errorf("Capture() error = %v, wantErr %v", captureErr, tt.wantErr)
			}
		})
	}

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&requestCount) == 0 {
		t.Error("Expected at least one request to the server")
	}
}
