package vigilant

import (
	"fmt"
)

// Logging functions are used to log messages, they are searchable in the Vigilant Dashboard.
// Only the information provided in the logs is sent to Vigilant.

// -------------------------- //
// --- Attributes Helpers --- //
// -------------------------- //

// AddGlobalAttributes adds global attributes to the global instance
func AddGlobalAttributes(attributes ...Attribute) {
	attrs, err := attributesToMap(attributes...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalInstance.addGlobalAttributes(attrs)
}

// ----------------------- //
// --- General Logging --- //
// ----------------------- //

// Log logs a message at the given level
//
// Use this function when you want to log a message at the given level.
//
// Example:
//
//	Log(LEVEL_INFO, "Hello, world!")
func Log(level LogLevel, message string) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(level, message, nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogError logs an error at the given level
//
// Use this function when you want to log an error.
//
// Example:
//
//	LogError("Failed to write to file")
func LogError(message string) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_ERROR, message, nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogWarn logs a warning at the given level
//
// Use this function when you want to log a warning.
//
// Example:
//
//	LogWarn("Failed to write to file")
func LogWarn(message string) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_WARN, message, nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogInfo logs an info message at the given level
//
// Use this function when you want to log an info message.
//
// Example:
//
//	LogInfo("Hello, world!")
func LogInfo(message string) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_INFO, message, nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogDebug logs a debug message at the given level
//
// Use this function when you want to log a debug message.
//
// Example:
//
//	LogDebug("Hello, world!")
func LogDebug(message string) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_DEBUG, message, nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogTrace logs a trace message at the given level
//
// Use this function when you want to log a trace message.
//
// Example:
//
//	LogTrace("Hello, world!")
func LogTrace(message string) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_TRACE, message, nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// ------------------------- //
// --- Formatted Logging --- //
// ------------------------- //

// LogErrorf logs an error at the given level
//
// Use this function when you want to log an error with a formatted message.
//
// Example:
// LogErrorf("Failed to %s", "do something")
func LogErrorf(template string, args ...any) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_ERROR, fmt.Sprintf(template, args...), nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogWarnf logs a warning at the given level
//
// Use this function when you want to log a warning with a formatted message.
//
// Example:
//
//	LogWarnf("Failed to %s", "do something")
func LogWarnf(template string, args ...any) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_WARN, fmt.Sprintf(template, args...), nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogInfof logs an info message at the given level
//
// Use this function when you want to log an info message with a formatted message.
//
// Example:
//
//	LogInfof("Failed to %s", "do something")
func LogInfof(template string, args ...any) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_INFO, fmt.Sprintf(template, args...), nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogDebugf logs a debug message at the given level
//
// Use this function when you want to log a debug message with a formatted message.
//
// Example:
//
//	LogDebugf("Failed to %s", "do something")
func LogDebugf(template string, args ...any) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_DEBUG, fmt.Sprintf(template, args...), nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogTracef logs a trace message at the given level
//
// Use this function when you want to log a trace message with a formatted message.
//
// Example:
//
//	LogTracef("Failed to %s", "do something")
func LogTracef(template string, args ...any) {
	if gateNilGlobalInstance() {
		return
	}

	log := createLogMessage(LEVEL_TRACE, fmt.Sprintf(template, args...), nil)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// ------------------------------- //
// --- Typed Attribute Logging --- //
// ------------------------------- //

// LogErrort logs an error at the given level with typed attributes
//
// Use this function when you want to log an error with typed attributes.
//
// Example:
//
//	LogErrort("Failed to write to file", "file", "example.txt", "error", "some error")
func LogErrort(message string, attributes ...Attribute) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := attributesToMap(attributes...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_ERROR, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogWarnt logs a warning at the given level with typed attributes
//
// Use this function when you want to log a warning with typed attributes.
//
// Example:
//
//	LogWarnt("Failed to write to file", "file", "example.txt", "error", "some error")
func LogWarnt(message string, attributes ...Attribute) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := attributesToMap(attributes...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_WARN, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogInfot logs an info message at the given level with typed attributes
//
// Use this function when you want to log an info message with typed attributes.
//
// Example:
//
//	LogInfot("Failed to write to file", "file", "example.txt", "error", "some error")
func LogInfot(message string, attributes ...Attribute) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := attributesToMap(attributes...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_INFO, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogDebugt logs a debug message at the given level with typed attributes
//
// Use this function when you want to log a debug message with typed attributes.
//
// Example:
//
//	LogDebugt("Failed to write to file", "file", "example.txt", "error", "some error")
func LogDebugt(message string, attributes ...Attribute) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := attributesToMap(attributes...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_DEBUG, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogTracet logs a trace message at the given level with typed attributes
//
// Use this function when you want to log a trace message with typed attributes.
//
// Example:
//
//	LogTracet("Failed to write to file", "file", "example.txt", "error", "some error")
func LogTracet(message string, attributes ...Attribute) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := attributesToMap(attributes...)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_TRACE, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// -------------------------------- //
// --- Free-form Attribute Logs --- //
// -------------------------------- //

// LogErrorw logs an error at the given level with key-value attributes
//
// Use this function when you want to log an error with key-value attributes.
//
// Example:
//
//	LogErrorw("Failed to write to file", "file", "example.txt", "error", "some error")
func LogErrorw(message string, keyVals ...any) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_ERROR, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogWarnw logs a warning at the given level with key-value attributes
//
// Use this function when you want to log a warning with key-value attributes.
//
// Example:
//
//	LogWarnw("Database query too long", "query", "SELECT * FROM users", "duration", "100ms")
func LogWarnw(message string, keyVals ...any) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_WARN, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogInfow logs an info message at the given level with key-value attributes
//
// Use this function when you want to log an info message with key-value attributes.
//
// Example:
//
//	LogInfow("User signup request", "email", "test@example.com")
func LogInfow(message string, keyVals ...any) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_INFO, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// LogDebugw logs a debug message at the given level with key-value attributes
//
// Use this function when you want to log a debug message with key-value attributes.
//
// Example:
//
//	LogDebugw("Timer tick", "time", "100ms")
func LogDebugw(message string, keyVals ...any) {
	if gateNilGlobalInstance() {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	log := createLogMessage(LEVEL_DEBUG, message, attrs)
	if log == nil {
		return
	}

	globalInstance.captureLog(log)
}

// writeLogPassthrough writes a log message to Vigilant
// this is an internal function that is used to write log messages to stdout
func writeLogPassthrough(level LogLevel, message string, attrs map[string]string) {
	switch level {
	case LEVEL_ERROR:
		if len(attrs) > 0 {
			fmt.Printf("[ERROR] %s %s\n", message, prettyPrintAttributes(attrs))
		} else {
			fmt.Printf("[ERROR] %s\n", message)
		}
	case LEVEL_WARN:
		if len(attrs) > 0 {
			fmt.Printf("[WARN] %s %s\n", message, prettyPrintAttributes(attrs))
		} else {
			fmt.Printf("[WARN] %s\n", message)
		}
	case LEVEL_INFO:
		if len(attrs) > 0 {
			fmt.Printf("[INFO] %s %s\n", message, prettyPrintAttributes(attrs))
		} else {
			fmt.Printf("[INFO] %s\n", message)
		}
	case LEVEL_DEBUG:
		if len(attrs) > 0 {
			fmt.Printf("[DEBUG] %s %s\n", message, prettyPrintAttributes(attrs))
		} else {
			fmt.Printf("[DEBUG] %s\n", message)
		}
	case LEVEL_TRACE:
		if len(attrs) > 0 {
			fmt.Printf("[TRACE] %s %s\n", message, prettyPrintAttributes(attrs))
		} else {
			fmt.Printf("[TRACE] %s\n", message)
		}
	}
}
