package vigilant

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// LoggerOptions are the options for the Logger
type LoggerOptions struct {
	otelLogger log.Logger
	name       string
	attributes []log.KeyValue
	url        string
	apiKey     string
}

// NewLoggerOptions creates a new LoggerOptions
func NewLoggerOptions(opts ...LoggerOption) *LoggerOptions {
	options := &LoggerOptions{
		otelLogger: nil,
		name:       "",
		attributes: []log.KeyValue{},
		url:        "",
		apiKey:     "",
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// LoggerOption is a function that configures the logger
type LoggerOption func(*LoggerOptions)

// WithService adds the service name to the logger
func WithName(name string) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.name = name
	}
}

// WithAttributes adds default attributes to all log messages
func WithAttributes(attrs ...log.KeyValue) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.attributes = append(opts.attributes, attrs...)
	}
}

// WithURL adds the URL to the logger
func WithURL(url string) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.url = url
	}
}

// WithAPIKey adds the API key to the logger
func WithAPIKey(apiKey string) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.apiKey = apiKey
	}
}

// WithOTELLogger adds the OTEL logger to the logger
func WithOTELLogger(otelLogger log.Logger) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.otelLogger = otelLogger
	}
}

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

// NewLogger creates a new Logger instance, falling back to noop logger on error
func NewLogger(
	opts *LoggerOptions,
) *Logger {
	otelLogger, err := getOtelLogger(opts)
	if err != nil {
		panic(err)
	}

	return &Logger{
		otelLogger: otelLogger,
		attributes: opts.attributes,
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
func (l *Logger) log(
	ctx context.Context,
	level LogLevel,
	message string,
	err error,
	attrs ...log.KeyValue,
) {
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

// newOtelLogger creates a new OpenTelemetry logger with OTLP export
func newOtelLogger(
	url string,
	name string,
) (log.Logger, error) {
	exporter, err := otlploggrpc.New(
		context.Background(),
		otlploggrpc.WithEndpoint(url),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	attrs := []attribute.KeyValue{}
	if name != "" {
		attrs = append(attrs, semconv.ServiceName(name))
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exporter),
		),
	)

	return provider.Logger(name), nil
}

// getOtelLogger creates a new OpenTelemetry logger with OTLP export
func getOtelLogger(
	opts *LoggerOptions,
) (log.Logger, error) {
	if opts.otelLogger != nil {
		return opts.otelLogger, nil
	}

	var name string = "example"
	if opts.name != "" {
		name = opts.name
	}

	var url string = "https://log.vigilant.run:4317"
	if opts.url != "" {
		url = opts.url
	}

	return newOtelLogger(url, name)
}
