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

## Usage (Event Capture)

```go
package main

import (
    "context"

    "github.com/vigilant-run/vigilant-golang"
)

func main() {
    // Create the event capture options
    eventCaptureOptions := vigilant.NewEventCaptureOptions(
        vigilant.WithEventCaptureURL("https://events.vigilant.run"),
        vigilant.WithEventCaptureToken("tk_1234567890"),
        vigilant.WithEventCaptureName("sample-app"),
    )

    // Create the event capture
    eventCapture := vigilant.NewEventCapture(eventCaptureOptions)

    // Capture an error
    err := errors.New("This is a test error")

    // Capture the error
    eventCapture.CaptureError(err)

    // Capture a message
    eventCapture.CaptureMessage("This is a test message")
}
```
