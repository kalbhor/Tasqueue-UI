.PHONY: build run clean test help dev install

# Build variables
BINARY_NAME=tasqueue-ui
BUILD_DIR=bin
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build      - Build the application binary"
	@echo "  run        - Build and run the application"
	@echo "  dev        - Run with hot reload (requires air)"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install the binary to GOPATH/bin"
	@echo "  help       - Display this help message"

## build: Build the application binary
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/server
	@echo "Built: ${BUILD_DIR}/${BINARY_NAME}"

## run: Build and run the application
run: build
	@echo "Starting ${BINARY_NAME}..."
	@./${BUILD_DIR}/${BINARY_NAME}

## dev: Run with default settings (in-memory broker for development)
dev: build
	@echo "Starting ${BINARY_NAME} in development mode..."
	@./${BUILD_DIR}/${BINARY_NAME} -broker in-memory

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ${BUILD_DIR}
	@go clean
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## install: Install the binary to GOPATH/bin
install: build
	@echo "Installing ${BINARY_NAME} to ${GOPATH}/bin..."
	@cp ${BUILD_DIR}/${BINARY_NAME} ${GOPATH}/bin/
	@echo "Installation complete"

## deps: Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "Dependencies verified"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "golangci-lint not installed"

.DEFAULT_GOAL := help
