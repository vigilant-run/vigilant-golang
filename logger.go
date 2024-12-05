package vigilant

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/log"
)

// Logger wraps the OpenTelemetry logger
type Logger struct {
	otelLogger log.Logger
	attributes []log.KeyValue
}

// LogLevel represents the severity of the log message
type LogLevel string

const (
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
	DebugLevel LogLevel = "DEBUG"
)

// NewLogger creates a new Logger instance
func NewLogger(otelLogger log.Logger, opts ...LoggerOption) *Logger {
	l := &Logger{
		otelLogger: otelLogger,
		attributes: []log.KeyValue{},
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// LoggerOption is a function that configures the logger
type LoggerOption func(*Logger)

// WithAttributes adds default attributes to all log messages
func WithAttributes(attrs ...log.KeyValue) LoggerOption {
	return func(l *Logger) {
		l.attributes = append(l.attributes, attrs...)
	}
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(ctx context.Context, message string, attrs ...log.KeyValue) {
	l.log(ctx, DebugLevel, message, nil, attrs...)
}

// Debugf logs a formatted message at DEBUG level
func (l *Logger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, DebugLevel, fmt.Sprintf(format, args...), nil, l.attributes...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(ctx context.Context, message string, attrs ...log.KeyValue) {
	l.log(ctx, WarnLevel, message, nil, attrs...)
}

// Warnf logs a formatted message at WARN level
func (l *Logger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, WarnLevel, fmt.Sprintf(format, args...), nil, l.attributes...)
}

// Info logs a message at INFO level
func (l *Logger) Info(ctx context.Context, message string, attrs ...log.KeyValue) {
	l.log(ctx, InfoLevel, message, nil, attrs...)
}

// Infof logs a formatted message at INFO level
func (l *Logger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, InfoLevel, fmt.Sprintf(format, args...), nil, l.attributes...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(ctx context.Context, message string, err error, attrs ...log.KeyValue) {
	l.log(ctx, ErrorLevel, message, err, attrs...)
}

// Errorf logs a formatted message at ERROR level
func (l *Logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, ErrorLevel, fmt.Sprintf(format, args...), nil, l.attributes...)
}

// log handles the actual logging
func (l *Logger) log(ctx context.Context, level LogLevel, message string, err error, attrs ...log.KeyValue) {
	record := log.Record{}
	record.SetSeverity(getSeverity(level))
	record.SetBody(log.StringValue(message))
	record.SetTimestamp(time.Now())

	allAttrs := append(l.attributes, attrs...)
	record.AddAttributes(allAttrs...)

	if err != nil {
		record.AddAttributes(log.String("error", err.Error()))
	}

	l.otelLogger.Emit(ctx, record)
}

// getSeverity converts our LogLevel to OTEL severity
func getSeverity(level LogLevel) log.Severity {
	switch level {
	case InfoLevel:
		return log.SeverityInfo
	case WarnLevel:
		return log.SeverityWarn
	case ErrorLevel:
		return log.SeverityError
	case DebugLevel:
		return log.SeverityDebug
	default:
		return log.SeverityInfo
	}
}
