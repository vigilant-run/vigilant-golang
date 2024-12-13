package vigilant

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
)

type mockLogger struct {
	embedded.Logger
	records []log.Record
}

func (m *mockLogger) Emit(_ context.Context, record log.Record) {
	m.records = append(m.records, record)
}

func (m *mockLogger) Enabled(_ context.Context, _ log.EnabledParameters) bool {
	return true
}

func TestLogger(t *testing.T) {
	t.Run("Info logging", func(t *testing.T) {
		mock := &mockLogger{}
		opts := &LoggerOptions{
			otelLogger: mock,
		}
		logger := NewLogger(opts)

		logger.Info(context.Background(), "test message", NewAttribute("count", 1))

		if len(mock.records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(mock.records))
		}
		assertRecord(t, mock.records[0], "test message", log.SeverityInfo)
	})

	t.Run("Error logging", func(t *testing.T) {
		mock := &mockLogger{}
		opts := &LoggerOptions{
			otelLogger: mock,
		}
		logger := NewLogger(opts)
		testErr := errors.New("test error")

		logger.Error(context.Background(), "error message", testErr)

		if len(mock.records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(mock.records))
		}

		record := mock.records[0]
		assertRecord(t, record, "error message", log.SeverityError)

		if !findAttribute(record, "error", testErr.Error()) {
			t.Error("error attribute not found in log record")
		}
	})

	t.Run("Default attributes", func(t *testing.T) {
		mock := &mockLogger{}
		opts := &LoggerOptions{
			otelLogger: mock,
			attributes: []Attribute{NewAttribute("default", "value")},
		}
		logger := NewLogger(opts)

		logger.Info(context.Background(), "test")

		if len(mock.records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(mock.records))
		}

		if !findAttribute(mock.records[0], "default", "value") {
			t.Error("default attribute not found in log record")
		}
	})

	t.Run("Log levels", func(t *testing.T) {
		testCases := []struct {
			level    LogLevel
			expected log.Severity
		}{
			{InfoLevel, log.SeverityInfo},
			{WarnLevel, log.SeverityWarn},
			{ErrorLevel, log.SeverityError},
			{DebugLevel, log.SeverityDebug},
		}

		for _, tc := range testCases {
			t.Run(string(tc.level), func(t *testing.T) {
				if got := getSeverity(tc.level); got != tc.expected {
					t.Errorf("getSeverity(%v) = %v, want %v", tc.level, got, tc.expected)
				}
			})
		}
	})
}

func assertRecord(t *testing.T, record log.Record, expectedMsg string, expectedSeverity log.Severity) {
	t.Helper()
	if record.Body().AsString() != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, record.Body().AsString())
	}
	if record.Severity() != expectedSeverity {
		t.Errorf("expected severity %v, got %v", expectedSeverity, record.Severity())
	}
}

func findAttribute(record log.Record, key, value string) bool {
	var found bool
	record.WalkAttributes(func(kv log.KeyValue) bool {
		if kv.Key == key && kv.Value.AsString() == value {
			found = true
			return false
		}
		return true
	})
	return found
}
