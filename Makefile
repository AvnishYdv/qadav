# Makefile for Qadav project
# Compatible with GNU make and PowerShell on Windows

BINARY     = bin/qyn.exe
LSP_BINARY = bin/qyn-lsp.exe
GOFLAGS    =
SHELL      = powershell.exe
.SHELLFLAGS = -NoProfile -Command

.PHONY: all build build-lsp build-all doctor test test-race test-race-linux test-update lint fmt clean ci

all: build

build:
	if (-not (Test-Path 'bin')) { New-Item -ItemType Directory -Path 'bin' | Out-Null }
	go build $(GOFLAGS) -o $(BINARY) ./cmd/qyn
	Write-Host "Build complete: $(BINARY)"

build-lsp:
	if (-not (Test-Path 'bin')) { New-Item -ItemType Directory -Path 'bin' | Out-Null }
	go build $(GOFLAGS) -o $(LSP_BINARY) ./cmd/qyn-lsp
	Write-Host "Build complete: $(LSP_BINARY)"

build-all: build build-lsp

doctor:
	go run ./cmd/qyn doctor

test:
	go test ./... -v

test-race:
	go test ./...

test-race-linux:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-update:
	go test ./... -args -update

lint:
	golangci-lint run

fmt:
	go fmt ./...

clean:
	go clean -cache -testcache
	if (Test-Path 'bin') { Remove-Item -Recurse -Force 'bin' }
	if (Test-Path 'coverage.out') { Remove-Item -Force 'coverage.out' }
	Write-Host "Clean complete."

ci: fmt lint test-race-linux
