# Vigilant Go Agent

This is the Go SDK for Vigilant.

You can learn more about Vigilant at [website](https://vigilant.run) or [docs](https://docs.vigilant.run).

## Installation

```bash
go get github.com/vigilant-run/vigilant-golang/v2
```

## Setup

You setup Vigilant by initializing the library at the start of your application.

```go
package main

import (
  "github.com/vigilant-run/vigilant-golang/v2"
)

func main() {
  // Create the vigilant config
  config := vigilant.NewConfigBuilder().
    WithName("backend").
    WithToken("tk_1234567890"). // Generate this from the Vigilant dashboard
    Build()

  // Initialize Vigilant
  vigilant.Init(config)

  // Shutdown Vigilant when the program exits
  defer vigilant.Shutdown()
}
```

## Logs 

You can learn more about logging in Vigilant in the [docs](https://docs.vigilant.run/logs).

```go
import (
  "github.com/vigilant-run/vigilant-golang/v2"
)

func function() {
  // Log a message
  vigilant.LogInfo("Hello, World!")
  vigilant.LogError("An error occurred")
  vigilant.LogWarn("A warning occurred")
  vigilant.LogDebug("A debug message")
  vigilant.LogTrace("A trace message")

  // Log a formatted message
  vigilant.LogInfof("Hello, %s!", "World")
  vigilant.LogErrorf("An error occurred: %s", "error")
  vigilant.LogWarnf("A warning occurred: %s", "warning")
  vigilant.LogDebugf("A debug message: %s", "debug")
  vigilant.LogTracef("A trace message: %s", "trace")

  // Log with typed attributes
  vigilant.LogErrort("An error occurred", vigilant.String("error", "some error"))
  vigilant.LogWarnt("A warning occurred", vigilant.Int("warning", 123))
  vigilant.LogInfot("A info message", vigilant.Bool("info", true))
  vigilant.LogDebugt("A debug message", vigilant.Float64("debug", 123.456))
  vigilant.LogTracet("A trace message", vigilant.Time("trace", time.Now()))

  // Log with key-value attributes (automatically converted to attribute pairs)
  vigilant.LogErrorw("An error occurred", "error", "some error")
  vigilant.LogWarnw("A warning occurred", "warning", "some warning")
  vigilant.LogInfow("A info message", "info", "some info")
  vigilant.LogDebugw("A debug message", "debug", "some debug")
  vigilant.LogTracew("A trace message", "trace", "some trace")
}
```

## Metrics

You can learn more about metrics in Vigilant in the [docs](https://docs.vigilant.run/metrics).

```go
import (
  "github.com/vigilant-run/vigilant-golang/v2"
)

func function() {
  // Create a counter metric 
  vigilant.MetricCounter("user_login_count", 1.0)
  vigilant.MetricCounter("user_login_count", 1.0, vigilant.Tag("env", "production"))

  // Create a gauge metric
  vigilant.MetricGauge("active_users", 1.0, vigilant.GaugeModeSet)
  vigilant.MetricGauge("active_users", 1.0, vigilant.GaugeModeSet, vigilant.Tag("env", "production"))

  // Create a histogram metric
  vigilant.MetricHistogram("request_duration", 123.4)
  vigilant.MetricHistogram("request_duration", 123.4, vigilant.Tag("env", "production"))
}
```
