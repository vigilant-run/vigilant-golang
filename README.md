# Vigilant Go SDK
This is the Go SDK for the Vigilant logging platform.

## Installation

```bash
go get github.com/vigilant-go/vigilant-golang
```

## Usage

```go
import (
	"github.com/vigilant-run/vigilant-golang"
)

func main() {
	loggerOptions := vigilant.NewLoggerOptions(
		vigilant.WithURL("https://log.vigilant.run:4317"),
		vigilant.WithToken("tk_1234567890"),
		vigilant.WithName("sample-app"),
	)

	logger := vigilant.NewLogger(loggerOptions)

	logger.Info(context.Background(), "Hello, World!")
}
```
