# Lint the application
lint:
	@echo "Linting..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run --tests=false --disable-all --timeout=2m -p error

# Generate mocks
mock:
	@echo "Generating mocks..."
	@go install github.com/vektra/mockery/v2@latest
	@mockery --config ./.mockery.yaml

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
.PHONY: help mock lint test
