package vigilant

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Error capture functions are used to capture errors, they are viewable in the Vigilant Dashboard.
// The Vigilant agent collects a bunch of metadata about the error, such as the stack trace, function name, file name, and line number.
// This data is used to provide more context about the error in the Vigilant Dashboard.

// ----------------------- //
// --- General Errors --- //
// ----------------------- //

// CaptureError captures an error and sends it to the agent
// Example:
// err := db.Query(...)
// CaptureError(err)
func CaptureError(err error) {
	if globalAgent == nil || err == nil {
		return
	}

	location := getLocation(2)
	details := getDetails(err)

	globalAgent.sendError(err, location, details, nil)
}

// CaptureErrorw captures an error and sends it to the agent with attributes
// Example:
// err := db.Query(...)
// CaptureErrorw(err, "db", "postgres")
func CaptureErrorw(err error, keyVals ...any) {
	if globalAgent == nil || err == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	location := getLocation(2)
	details := getDetails(err)

	globalAgent.sendError(err, location, details, attrs)
}

// CaptureErrort captures an error message and sends it to the agent with typed attributes
// Example:
// err := db.Query(...)
// CaptureErrort(err, vigilant.String("db", "postgres"))
func CaptureErrort(message string, attrs map[string]string) {
	if globalAgent == nil || message == "" {
		return
	}

	location := getLocation(2)
	details := getDetails(errors.New(message))
	globalAgent.sendError(errors.New(message), location, details, attrs)
}

// ----------------------- //
// --- Wrapped Errors --- //
// ----------------------- //
// CaptureWrappedError wraps an error and sends it to the agent
// Example:
// err := db.Query(...)
// CaptureWrappedError(err)
func CaptureWrappedError(message string, err error) {
	if globalAgent == nil || err == nil {
		return
	}

	location := getLocation(2)
	details := getDetails(err)

	globalAgent.sendError(err, location, details, nil)
}

// CaptureWrappedErrorw wraps an error and sends it to the agent with attributes
// Example:
// err := db.Query(...)
// CaptureWrappedErrorw(err, "db", "postgres")
func CaptureWrappedErrorw(message string, err error, keyVals ...any) {
	if globalAgent == nil || err == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	location := getLocation(2)
	details := getDetails(err)

	globalAgent.sendError(err, location, details, attrs)
}

// CaptureWrappedErrort wraps an error message and sends it to the agent with typed attributes
// Example:
// err := db.Query(...)
// CaptureWrappedErrort(err, vigilant.String("db", "postgres"))
func CaptureWrappedErrort(message string, err error, attrs map[string]string) {
	if globalAgent == nil || err == nil {
		return
	}

	location := getLocation(2)
	details := getDetails(err)

	globalAgent.sendError(err, location, details, attrs)
}

// ----------------------- //
// --- Error Message --- //
// ----------------------- //

// CaptureMessage captures an error message and sends it to the agent
// Example:
// CaptureMessage("failed to write to file")
func CaptureMessage(message string) {
	if globalAgent == nil || message == "" {
		return
	}

	capturedErr := errors.New(message)
	location := getLocation(2)
	details := getDetails(capturedErr)

	globalAgent.sendError(capturedErr, location, details, nil)
}

// CaptureMessagef captures an error message and sends it to the agent
// Example:
// CaptureMessagef("failed to write to file: %s", "file.txt")
func CaptureMessagef(template string, args ...any) {
	if globalAgent == nil || template == "" {
		return
	}

	capturedErr := fmt.Errorf(template, args...)
	location := getLocation(2)
	details := getDetails(capturedErr)

	globalAgent.sendError(capturedErr, location, details, nil)
}

// CaptureMessagew captures an error message and sends it to the agent with attributes
// Example:
// CaptureMessagew("failed to write to file", "file.txt", "db", "postgres")
func CaptureMessagew(template string, args ...any) {
	if globalAgent == nil || template == "" {
		return
	}

	attrs, err := keyValsToMap(args...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	capturedErr := fmt.Errorf(template, args...)
	location := getLocation(2)
	details := getDetails(capturedErr)

	globalAgent.sendError(capturedErr, location, details, attrs)
}

// CaptureMessaget captures an error message and sends it to the agent with attributes
// Example:
// -- CaptureMessaget("failed to write to file", vigilant.String("db", "postgres"))
func CaptureMessaget(message string, fields ...Field) {
	if globalAgent == nil || message == "" {
		return
	}
	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting fields: %v\n", err)
		return
	}
	capturedErr := errors.New(message)
	location := getLocation(2)
	details := getDetails(capturedErr)

	globalAgent.sendError(capturedErr, location, details, attrs)
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

		sb.WriteString(fmt.Sprintf("\tFile \"%s\", line %d, in %s\n", frame.File, frame.Line, funcName))
		if !more {
			break
		}
	}

	return sb.String()
}

// writeErrorPassthrough writes an error message to the agent
// this is an internal function that is used to write error messages to stdout
func writeErrorPassthrough(err error, attrs map[string]string) {
	if err == nil {
		return
	}
	if len(attrs) > 0 {
		formattedAttrs := prettyPrintAttributes(attrs)
		fmt.Printf("[ERROR] %s %s\n", err.Error(), formattedAttrs)
	} else {
		fmt.Printf("[ERROR] %s\n", err.Error())
	}
}
