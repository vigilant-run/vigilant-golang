# Vigilant Go SDK

This is the Go SDK for the Vigilant platform. The logger is a wrapper around the [OpenTelemetry](https://opentelemetry.io/) SDK. 
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

    // Shutdown the logger
    logger.Shutdown(context.Background())
}
```