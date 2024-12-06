# Vigilant Go SDK

This is the Go SDK for the Vigilant logging platform. It is a wrapper around the [OpenTelemetry](https://opentelemetry.io/) SDK. It allows to easily use the Vigilant logging platform in your Go applications without Vendor Lock-In.

## Installation

```bash
go get github.com/vigilant-go/vigilant-golang
```

## Usage

```go
package main

import (
    "context"

    "github.com/vigilant-run/vigilant-golang"
)

func main() {
    loggerOptions := vigilant.NewLoggerOptions(
        vigilant.WithURL("log.vigilant.run:4317"),
        vigilant.WithToken("tk_1234567890"),
        vigilant.WithName("sample-app"),
        vigilant.WithPassthrough(),
    )

    logger := vigilant.NewLogger(loggerOptions)

    logger.Info(context.Background(), "Hello, World!")
}
```
