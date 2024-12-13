# Vigilant Go SDK

This is the Go SDK for the Vigilant platform. The logger is a wrapper around the [OpenTelemetry](https://opentelemetry.io/) SDK. The error handler is a custom error handler that sends errors to the Vigilant platform. Together they allow you to correlate errors with logs in the Vigilant platform.

## Installation

```bash
go get github.com/vigilant-run/vigilant-golang
```

## Usage (Logger)

```go
package main

import (
    "context"

    "github.com/vigilant-run/vigilant-golang"
)

func main() {
    // Create the logger options
    loggerOptions := vigilant.NewLoggerOptions(
        vigilant.WithLoggerURL("log.vigilant.run:4317"),
        vigilant.WithLoggerToken("tk_1234567890"),
        vigilant.WithLoggerName("sample-app"),
        vigilant.WithLoggerPassthrough(),
    )

    // Create the logger
    logger := vigilant.NewLogger(loggerOptions)

    // Log a message
    logger.Info(context.Background(), "Hello, World!")
}
```

## Usage (Error Handler)

```go
package main

import (
    "context"

    "github.com/vigilant-run/vigilant-golang"
)

func main() {
    // Create the error handler options
    errorHandlerOptions := vigilant.NewErrorHandlerOptions(
        vigilant.WithErrorHandlerURL("https://errors.vigilant.run"),
        vigilant.WithErrorHandlerToken("tk_1234567890"),
        vigilant.WithErrorHandlerName("sample-app"),
    )

    // Create the error handler
    errorHandler := vigilant.NewErrorHandler(errorHandlerOptions)

    // Capture an error
    err := errors.New("This is a test error")

    // Capture the error
    errorHandler.Capture(context.Background(), err)
}
```
