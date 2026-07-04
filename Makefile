.PHONY: all build test clean run

# Default target
all: build test

# Build the project
build:
	@echo "Building..."
	go build -o bin/ ./cmd/...

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/

# Example run target (update with your primary binary)
run:
	@echo "Running application..."
	go run ./cmd/...
