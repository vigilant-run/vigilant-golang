# Vigilant Go Agent

This is the Go SDK for Vigilant.

## Installation

```bash
go get github.com/vigilant-run/vigilant-golang/v2
```

## Usage

```go
package main

import (
  "github.com/vigilant-run/vigilant-golang"
)

func main() {
  // Create the vigilant config
  config := vigilant.NewVigilantConfigBuilder().
    WithName("backend").
    WithToken("tk_1234567890").
    Build()

  // Initialize vigilant
  vigilant.Init(config)

  // Log a message
  vigilant.LogInfo("Hello, World!")

  // Shutdown vigilant
  vigilant.Shutdown()
}
```