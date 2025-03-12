# Lint the application
lint:
	@echo "Linting..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run --tests=false --disable-all --timeout=2m -p error

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

 help:
	@echo "Available commands:"
	@echo "  mock - Generate mocks"
	@echo "  lint - Lint the library"
	@echo "  test - Test the library"

.DEFAULT_GOAL := help
.PHONY: help lint test
