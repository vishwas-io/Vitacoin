#!/usr/bin/make -f

###############################################################################
###                                VITACOIN                                 ###
###                    Production-Grade Makefile                            ###
###############################################################################

BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
COMMIT := $(shell git log -1 --format='%H' 2>/dev/null)
BUILD_DIR ?= build
LEDGER_ENABLED ?= true
TM_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
BINDIR ?= $(GOPATH)/bin

# Binary names
VITACOIND = vitacoind

# Go version
GO_VERSION := 1.21

# Build flags
export GO111MODULE = on
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags 2>/dev/null || echo "v0.1.0") | sed 's/^v//')
TMVERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
COMMIT := $(shell git log -1 --format='%H' 2>/dev/null)

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=vitacoin \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(VITACOIND) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=$(BUILD_TAGS) \
	-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TMVERSION)

BUILD_FLAGS := -ldflags '$(ldflags)' -tags '$(BUILD_TAGS)'

###############################################################################
###                              Build Commands                             ###
###############################################################################

# Note: go.mod is at parent directory (../go.mod), build from there
PROJECT_ROOT := $(shell cd .. && pwd)

all: lint test build

build:
	@echo "🔨 Building vitacoind..."
	cd "$(PROJECT_ROOT)" && go build -mod=readonly $(BUILD_FLAGS) -o "vitacoin/$(BUILD_DIR)/$(VITACOIND)" ./vitacoin/cmd/vitacoind
	@echo "✅ Build complete: $(BUILD_DIR)/$(VITACOIND)"

build-linux:
	@echo "🐧 Building for Linux..."
	cd "$(PROJECT_ROOT)" && LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o "vitacoin/vitacoin/$(BUILD_DIR)/$(VITACOIND)" ./vitacoin/vitacoin/cmd/vitacoind
	@echo "✅ Linux build complete"

install:
	@echo "📦 Installing vitacoind..."
	cd "$(PROJECT_ROOT)" && go install -mod=readonly $(BUILD_FLAGS) ./vitacoin/vitacoin/cmd/vitacoind
	@echo "✅ Installed to $(GOPATH)/bin/$(VITACOIND)"

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf vendor/
	@echo "✅ Clean complete"

###############################################################################
###                           Protocol Buffers                              ###
###############################################################################

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "🔧 Generating protobuf files..."
	@cd $(PROJECT_ROOT) && ./vitacoin/scripts/protocgen.sh
	@echo "✅ Proto generation complete"

proto-format:
	@echo "📝 Formatting proto files..."
	@find proto -name "*.proto" -exec clang-format -i {} \; 2>/dev/null || echo "⚠️  clang-format not installed, skipping"
	@echo "✅ Proto formatting complete"

proto-lint:
	@echo "🔍 Linting proto files..."
	@buf lint --error-format=json
	@echo "✅ Proto lint complete"

proto-update-deps:
	@echo "⬆️  Updating proto dependencies..."
	@buf mod update
	@echo "✅ Proto dependencies updated"

###############################################################################
###                              Testing                                    ###
###############################################################################

test: test-unit

test-all: test-unit test-race test-cover

test-unit:
	@echo "🧪 Running unit tests..."
	@cd $(PROJECT_ROOT) && go test -mod=readonly -v ./vitacoin/vitacoin/...
	@echo "✅ Unit tests complete"

test-race:
	@echo "🏁 Running race condition tests..."
	@cd $(PROJECT_ROOT) && go test -mod=readonly -race -v ./vitacoin/vitacoin/...
	@echo "✅ Race tests complete"

test-cover:
	@echo "📊 Running tests with coverage..."
	@cd $(PROJECT_ROOT) && go test -mod=readonly -coverprofile=vitacoin/vitacoin/coverage.out -covermode=atomic ./vitacoin/vitacoin/...
	@cd $(PROJECT_ROOT) && go tool cover -html=vitacoin/vitacoin/coverage.out -o vitacoin/vitacoin/coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-sim-nondeterminism:
	@echo "🎲 Running non-determinism simulation..."
	@cd $(PROJECT_ROOT) && go test -mod=readonly $(PACKAGES_SIMTEST) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h
	@echo "✅ Non-determinism test complete"

test-sim-import-export:
	@echo "📤 Running import/export simulation..."
	@cd $(PROJECT_ROOT) && go test -mod=readonly $(PACKAGES_SIMTEST) -run TestAppImportExport -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h
	@echo "✅ Import/export test complete"

###############################################################################
###                              Linting                                    ###
###############################################################################

lint:
	@echo "🔍 Running linters..."
	@cd $(PROJECT_ROOT) && golangci-lint run --config vitacoin/.golangci.yml ./vitacoin/vitacoin/...
	@echo "✅ Lint complete"

lint-fix:
	@echo "🔧 Running linters with auto-fix..."
	@cd $(PROJECT_ROOT) && golangci-lint run --config vitacoin/.golangci.yml --fix ./vitacoin/vitacoin/...
	@echo "✅ Lint fix complete"

format:
	@echo "📝 Formatting Go code..."
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "*vendor*" -not -path "*pb.go" | xargs gofmt -w -s
	@find . -name '*.go' -type f -not -path "*.git*" -not -path "*vendor*" -not -path "*pb.go" | xargs goimports -w -local github.com/vitacoin/vitacoin
	@echo "✅ Format complete"

###############################################################################
###                              Dependencies                               ###
###############################################################################

go-mod-verify:
	@echo "📥 Ensuring dependencies have not been modified..."
	@cd $(PROJECT_ROOT) && go mod verify

go-mod-tidy:
	@echo "🧹 Tidying go.mod..."
	@cd $(PROJECT_ROOT) && go mod tidy
	@echo "✅ go.mod tidied"

go-mod-download:
	@echo "⬇️  Downloading dependencies..."
	@cd $(PROJECT_ROOT) && go mod download
	@echo "✅ Dependencies downloaded"

###############################################################################
###                              Docker                                     ###
###############################################################################

docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t vitacoin:latest -f Dockerfile .
	@echo "✅ Docker build complete"

docker-run:
	@echo "🐳 Running Docker container..."
	@docker run -it --rm vitacoin:latest
	@echo "✅ Docker run complete"

###############################################################################
###                           Testnet Commands                              ###
###############################################################################

init-testnet:
	@echo "🌐 Initializing testnet..."
	@./scripts/init-testnet.sh
	@echo "✅ Testnet initialized"

start-testnet:
	@echo "🚀 Starting testnet..."
	@$(BUILD_DIR)/$(VITACOIND) start --home ./testnet/node0
	@echo "✅ Testnet started"

###############################################################################
###                               Utilities                                 ###
###############################################################################

.PHONY: all build build-linux install clean \
	proto-all proto-gen proto-format proto-lint proto-update-deps \
	test test-all test-unit test-race test-cover test-sim-nondeterminism test-sim-import-export \
	lint lint-fix format \
	go.sum go-mod-tidy \
	docker-build docker-run \
	init-testnet start-testnet