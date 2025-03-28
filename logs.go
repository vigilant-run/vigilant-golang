package vigilant

import (
	"fmt"
)

// Logging functions are used to log messages, they are searchable in the Vigilant Dashboard.
// Only the information provided in the logs is sent to Vigilant.

// ----------------------- //
// --- General Logging --- //
// ----------------------- //

// Log logs a message at the given level
// Example:
// Log(LEVEL_INFO, "Hello, world!")
func Log(level LogLevel, message string) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(level, message, nil)
}

// LogError logs an error at the given level
// Example:
// LogError("Failed to write to file")
func LogError(message string) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_ERROR, message, nil)
}

// LogWarn logs a warning at the given level
// Example:
// LogWarn("Failed to write to file")
func LogWarn(message string) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_WARN, message, nil)
}

// LogInfo logs an info message at the given level
// Example:
// LogInfo("Hello, world!")
func LogInfo(message string) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_INFO, message, nil)
}

// LogDebug logs a debug message at the given level
// Example:
// LogDebug("Hello, world!")
func LogDebug(message string) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_DEBUG, message, nil)
}

// LogTrace logs a trace message at the given level
// Example:
// LogTrace("Hello, world!")
func LogTrace(message string) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_TRACE, message, nil)
}

// ------------------------- //
// --- Formatted Logging --- //
// ------------------------- //

// LogErrorf logs an error at the given level
// Example:
// LogErrorf("Failed to %s", "do something")
func LogErrorf(template string, args ...any) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_ERROR, fmt.Sprintf(template, args...), nil)
}

// LogWarnf logs a warning at the given level
// Example:
// LogWarnf("Failed to %s", "do something")
func LogWarnf(template string, args ...any) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_WARN, fmt.Sprintf(template, args...), nil)
}

// LogInfof logs an info message at the given level
// Example:
// LogInfof("Failed to %s", "do something")
func LogInfof(template string, args ...any) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_INFO, fmt.Sprintf(template, args...), nil)
}

// LogDebugf logs a debug message at the given level
// Example:
// LogDebugf("Failed to %s", "do something")
func LogDebugf(template string, args ...any) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_DEBUG, fmt.Sprintf(template, args...), nil)
}

// LogTracef logs a trace message at the given level
// Example:
// LogTracef("Failed to %s", "do something")
func LogTracef(template string, args ...any) {
	if globalAgent == nil {
		return
	}

	globalAgent.sendLog(LEVEL_TRACE, fmt.Sprintf(template, args...), nil)
}

// ----------------------------------- //
// --- Typed Attribute Logging --- //
// ----------------------------------- //

// LogErrort logs an error at the given level with attributes
// Example:
// LogErrort("Failed to write to file", "file", "example.txt", "error", "some error")
func LogErrort(message string, fields ...Field) {
	if globalAgent == nil {
		return
	}

	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting fields: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_ERROR, message, attrs)
}

// LogWarnt logs a warning at the given level with attributes
// Example:
// LogWarnt("Failed to write to file", "file", "example.txt", "error", "some error")
func LogWarnt(message string, fields ...Field) {
	if globalAgent == nil {
		return
	}

	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting fields: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_WARN, message, attrs)
}

// LogInfot logs an info message at the given level with attributes
// Example:
// LogInfot("Failed to write to file", "file", "example.txt", "error", "some error")
func LogInfot(message string, fields ...Field) {
	if globalAgent == nil {
		return
	}

	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting fields: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_INFO, message, attrs)
}

// LogDebugt logs a debug message at the given level with attributes
// Example:
// LogDebugt("Failed to write to file", "file", "example.txt", "error", "some error")
func LogDebugt(message string, fields ...Field) {
	if globalAgent == nil {
		return
	}

	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting fields: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_DEBUG, message, attrs)
}

// LogTracet logs a trace message at the given level with attributes
// Example:
// LogTracet("Failed to write to file", "file", "example.txt", "error", "some error")
func LogTracet(message string, fields ...Field) {
	if globalAgent == nil {
		return
	}

	attrs, err := fieldsToMap(fields...)
	if err != nil {
		fmt.Printf("error formatting fields: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_TRACE, message, attrs)
}

// ----------------------------------- //
// --- Free-form Attribute Logging --- //
// ----------------------------------- //

// LogErrorw logs an error at the given level with attributes
// Example:
// LogErrorw("Failed to write to file", "file", "example.txt", "error", "some error")
func LogErrorw(message string, keyVals ...any) {
	if globalAgent == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_ERROR, message, attrs)
}

// LogWarnw logs a warning at the given level with attributes
// Example:
// LogWarnw("Database query too long", "query", "SELECT * FROM users", "duration", "100ms")
func LogWarnw(message string, keyVals ...any) {
	if globalAgent == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_WARN, message, attrs)
}

// LogInfow logs an info message at the given level with attributes
// Example:
// LogInfow("User signup request", "email", "test@example.com")
func LogInfow(message string, keyVals ...any) {
	if globalAgent == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_INFO, message, attrs)
}

// LogDebugw logs a debug message at the given level with attributes
// Example:
// LogDebugw("Timer tick", "time", "100ms")
func LogDebugw(message string, keyVals ...any) {
	if globalAgent == nil {
		return
	}

	attrs, err := keyValsToMap(keyVals)
	if err != nil {
		fmt.Printf("error formatting attributes: %v\n", err)
		return
	}

	globalAgent.sendLog(LEVEL_DEBUG, message, attrs)
}

// writeLogPassthrough writes a log message to the agent
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
