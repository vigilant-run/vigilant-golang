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
  "github.com/vigilant-run/vigilant-golang"
)

func main() {
  // Create the logger
  config := vigilant.NewLoggerConfigBuilder().
    WithName("sample-app").
    WithToken("tk_1234567890").
    Build()

  // Initialize the logger
  vigilant.InitLogger(config)

  // Log a message
  vigilant.LogInfo("Hello, World!")

  // Shutdown the logger
  vigilant.ShutdownLogger()
}
```