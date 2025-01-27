# Vigilant Go SDK

This is the Go SDK for the Vigilant platform.
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
  // Create the logger
  logger := vigilant.NewLoggerBuilder().
    WithName("sample-app").
    WithEndpoint("ingress.vigilant.run").
    WithToken("tk_1234567890").
    Build()

  // Log a message
  logger.Info("Hello, World!")

  // Shutdown the logger
  logger.Shutdown()
}
```