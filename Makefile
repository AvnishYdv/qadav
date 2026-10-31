# Makefile for Qadav project

# Variables
BINARY=bin/qyn
LSP_BINARY=bin/qyn-lsp

.PHONY: all build build-lsp build-all test test-race test-update lint fmt clean ci

all: build

build:
	@echo "Building..."
	@go build -o $(BINARY) ./cmd/qyn

build-lsp:
	@echo "Building LSP..."
	@go build -o $(LSP_BINARY) ./cmd/qyn-lsp

build-all: build build-lsp

test:
	@echo "Running tests..."
	@go test ./... -v

test-race:
	@go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-update:
	@go test -update ./...

lint:
	@golangci-lint run

fmt:
	@go fmt ./...
	@goimports -w .

clean:
	@echo "Cleaning..."
	@go clean -cache -testcache
	@rm -rf $(BINARY) $(LSP_BINARY) bin coverage.out 2>/dev/null || true
	@powershell -Command "if (Test-Path '$(BINARY)') { Remove-Item -Recurse -Force '$(BINARY)' }; if (Test-Path '$(LSP_BINARY)') { Remove-Item -Recurse -Force '$(LSP_BINARY)' }; if (Test-Path 'bin') { Remove-Item -Recurse -Force 'bin' }; if (Test-Path 'coverage.out') { Remove-Item -Force 'coverage.out' }"

ci: fmt lint test-race
