# Vigilant Go Agent

This is the Go Agent for the Vigilant platform.

## Installation

```bash
go get github.com/vigilant-run/vigilant-golang
```

## Usage

```go
package main

import (
  "github.com/vigilant-run/vigilant-golang"
)

func main() {
  // Create the agent config
  config := vigilant.NewAgentConfigBuilder().
    WithName("backend").
    WithToken("tk_1234567890").
    Build()

  // Initialize the agent
  vigilant.Init(config)

  // Log a message
  vigilant.LogInfo("Hello, World!")

  // Send an alert
  vigilant.SendAlert("Something went wrong")

  // Emit a metric
  vigilant.EmitMetric("my_metric", 1.0)

  // Shutdown the agent
  vigilant.Shutdown()
}
```