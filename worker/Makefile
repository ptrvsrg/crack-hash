# Build the application
ARTIFACT_VERSION ?= 0.0.0-local
build:
	@echo "Building..."
	@go build \
	-o ./bin/worker \
	-installsuffix "static" \
	-tags "" \
	-ldflags " \
	-X github.com/ptrvsrg/crack-hash/worker/internal/version.AppVersion=$(ARTIFACT_VERSION) \
	-X github.com/ptrvsrg/crack-hash/worker/internal/version.GoVersion=$(shell go version | cut -d " " -f 3) \
	-X github.com/ptrvsrg/crack-hash/worker/internal/version.Platform=$(shell go env GOOS)/$(shell go env GOARCH)" \
	./cmd/cli

# Build the docker image
build-image:
	@echo "Building image..."
	@docker build -t ptrvsrg/crack-hash-worker:$(ARTIFACT_VERSION) -f Dockerfile ..

# Run the application
COMMAND ?= server
run:
	@echo "Running..."
	@go run ./cmd/cli $(COMMAND)

# Generate Swagger spec
swagger:
	@echo "Generating swagger..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init --parseDependency --generalInfo ./internal/transport/http/router.go --outputTypes go,yaml --output ./docs
	sed -i '' 's/github_com_ptrvsrg_crack-hash_worker_pkg_//g' ./docs/docs.go
	sed -i '' 's/github_com_ptrvsrg_crack-hash_worker_pkg_//g' ./docs/swagger.yaml

# Generate mocks
mock:
	@echo "Generating mocks..."
	@go install github.com/vektra/mockery/v2@latest
	@mockery --config ./.mockery.yaml

# Lint the application
lint:
	@echo "Linting..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run --tests=false --disable-all --timeout=2m -p error

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f bin

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

 help:
	@echo "Available commands:"
	@echo "  build   		- Build the application"
	@echo "  build-image 		- Build the docker image"
	@echo "  run     		- Run the application (set the COMMAND environment variable to change the command, default is 'server')"
	@echo "  swagger 		- Generate Swagger specification"
	@echo "  mock    		- Generate mocks"
	@echo "  lint    		- Lint the application"
	@echo "  test    		- Test the application"
	@echo "  clean   		- Clean the binary"
	@echo "  watch   		- Live Reload"

.DEFAULT_GOAL := help
.PHONY: help build build-image run swagger mock lint test clean watch
