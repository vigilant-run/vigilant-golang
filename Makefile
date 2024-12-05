# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=vigilant
BINARY_UNIX=$(BINARY_NAME)_unix

# All target
all: test build

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Install dependencies
deps:
	$(GOGET) -v ./...

# Format the code
fmt:
	$(GOCMD) fmt ./...

# Tidy the dependencies
tidy:
	$(GOCMD) mod tidy

# Help
help:
	@echo "Makefile commands:"
	@echo "  make all        - Run tests and build the project"
	@echo "  make build      - Build the project"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make deps       - Install dependencies"
	@echo "  make build-linux - Cross compile for Linux"
	@echo "  make run        - Run the application"
	@echo "  make fmt        - Format the code"
	@echo "  make lint       - Lint the code (requires golangci-lint)"
