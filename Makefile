# Simple Makefile for ccctx

# Variables
BINARY=ccctx
MAIN_GO=main.go

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -o ${BINARY} .

# Install the binary to GOPATH/bin
.PHONY: install
install:
	CGO_ENABLED=0 go install .

# Clean build artifacts
.PHONY: clean
clean:
	rm -f ${BINARY}

# Run tests
.PHONY: test
test:
	go test ./...

# Format source code
.PHONY: fmt
fmt:
	go fmt ./...

# Vet the source code
.PHONY: vet
vet:
	go vet ./...

# Tidy go.mod and go.sum
.PHONY: tidy
tidy:
	go mod tidy

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all     - Build the binary (default)"
	@echo "  build   - Build the binary"
	@echo "  install - Install the binary to GOPATH/bin"
	@echo "  clean   - Remove build artifacts"
	@echo "  test    - Run tests"
	@echo "  fmt     - Format source code"
	@echo "  vet     - Vet the source code"
	@echo "  tidy    - Tidy go.mod and go.sum"
	@echo "  help    - Show this help message"
