package vigilant

import (
	"context"
	"fmt"
	"runtime"
	"strings"
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
	otelLogger  log.Logger
	name        string
	attributes  []Attribute
	url         string
	token       string
	passthrough bool
	noop        bool
	insecure    bool
}

// NewLoggerOptions creates a new LoggerOptions
func NewLoggerOptions(opts ...LoggerOption) *LoggerOptions {
	options := &LoggerOptions{
		otelLogger:  nil,
		name:        "go-server",
		attributes:  []Attribute{},
		url:         "log.vigilant.run:4317",
		token:       "tk_1234567890",
		passthrough: false,
		insecure:    false,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// LoggerOption is a function that configures the logger
type LoggerOption func(*LoggerOptions)

// WithLoggerName adds the service name to the logger
func WithLoggerName(name string) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.name = name
	}
}

// WithLoggerAttributes adds default attributes to all log messages
func WithLoggerAttributes(attrs ...Attribute) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.attributes = append(opts.attributes, attrs...)
	}
}

// WithLoggerURL adds the URL to the logger
func WithLoggerURL(url string) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.url = url
	}
}

// WithLoggerToken adds the token to the logger
func WithLoggerToken(token string) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.token = token
	}
}

// WithLoggerOTELLogger adds the OTEL logger to the logger
func WithLoggerOTELLogger(otelLogger log.Logger) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.otelLogger = otelLogger
	}
}

// WithLoggerPassthrough also logs fmt.Println
func WithLoggerPassthrough() LoggerOption {
	return func(opts *LoggerOptions) {
		opts.passthrough = true
	}
}

// WithLoggerNoop disables the logger
func WithLoggerNoop() LoggerOption {
	return func(opts *LoggerOptions) {
		opts.noop = true
	}
}

// WithLoggerInsecure disables TLS verification
func WithLoggerInsecure() LoggerOption {
	return func(opts *LoggerOptions) {
		opts.insecure = true
	}
}

// Logger wraps the OpenTelemetry logger
type Logger struct {
	otelLogger  log.Logger
	attributes  []Attribute
	passthrough bool
	noop        bool
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
		otelLogger:  otelLogger,
		attributes:  opts.attributes,
		passthrough: opts.passthrough,
	}
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(ctx context.Context, message string, attrs ...Attribute) {
	if !l.noop {
		callerAttrs := getCallerAttrs()
		allAttrs := append(l.attributes, callerAttrs...)
		allAttrs = append(allAttrs, attrs...)
		l.log(ctx, DebugLevel, message, nil, allAttrs...)
	}
	if l.passthrough {
		fmt.Println(message)
	}
}

// Warn logs a message at WARN level
func (l *Logger) Warn(ctx context.Context, message string, attrs ...Attribute) {
	if !l.noop {
		callerAttrs := getCallerAttrs()
		allAttrs := append(l.attributes, callerAttrs...)
		allAttrs = append(allAttrs, attrs...)
		l.log(ctx, WarnLevel, message, nil, allAttrs...)
	}
	if l.passthrough {
		fmt.Println(message)
	}
}

// Info logs a message at INFO level
func (l *Logger) Info(ctx context.Context, message string, attrs ...Attribute) {
	if !l.noop {
		callerAttrs := getCallerAttrs()
		allAttrs := append(l.attributes, callerAttrs...)
		allAttrs = append(allAttrs, attrs...)
		l.log(ctx, InfoLevel, message, nil, allAttrs...)
	}
	if l.passthrough {
		fmt.Println(message)
	}
}

// Error logs a message at ERROR level
func (l *Logger) Error(ctx context.Context, message string, err error, attrs ...Attribute) {
	if !l.noop {
		callerAttrs := getCallerAttrs()
		allAttrs := append(l.attributes, callerAttrs...)
		allAttrs = append(allAttrs, attrs...)
		l.log(ctx, ErrorLevel, message, err, allAttrs...)
	}
	if l.passthrough {
		fmt.Println(message)
	}
}

// log handles the actual logging
func (l *Logger) log(
	ctx context.Context,
	level LogLevel,
	message string,
	err error,
	attrs ...Attribute,
) {
	record := log.Record{}
	record.SetSeverity(getSeverity(level))
	record.SetBody(log.StringValue(message))
	record.SetTimestamp(time.Now())

	allAttrs := append(l.attributes, attrs...)
	logAttrs := []log.KeyValue{}
	for _, attr := range allAttrs {
		logAttrs = append(logAttrs, attr.ToLogKV())
	}

	record.AddAttributes(logAttrs...)

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
	token string,
	name string,
	insecure bool,
) (log.Logger, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(url),
		otlploggrpc.WithHeaders(map[string]string{
			"x-vigilant-token": token,
		}),
	}
	if insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(
		context.Background(),
		opts...,
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

	var url string = "otel.vigilant.run:4317"
	if opts.url != "" {
		url = opts.url
	}

	var token string = "tk_1234567890"
	if opts.token != "" {
		token = opts.token
	}

	var insecure bool = false
	if opts.insecure {
		insecure = opts.insecure
	}

	return newOtelLogger(url, token, name, insecure)
}

// getCallerAttrs returns the caller attributes
func getCallerAttrs() []Attribute {
	file, line, fn := getCallerInfo()
	return []Attribute{
		NewAttribute("caller.file", file),
		NewAttribute("caller.line", line),
		NewAttribute("caller.function", fn),
	}
}

// getCallerInfo returns the caller information
func getCallerInfo() (string, int, string) {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "", 0, ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return file, line, ""
	}

	name := fn.Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}

	return file, line, name
}
