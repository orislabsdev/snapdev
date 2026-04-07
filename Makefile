# snapdev Makefile
# Provides common development tasks. Run `make help` for a summary.

BINARY      := snapdev
MODULE      := github.com/orislabsdev/snapdev
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     := -s -w \
               -X $(MODULE)/cmd.Version=$(VERSION) \
               -X $(MODULE)/cmd.Commit=$(COMMIT) \
               -X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)

OUTPUT_DIR  := bin
DIST_DIR    := dist

PLATFORMS   := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build install clean test test-race lint vet fmt tidy deps build-all release snapshot help

##@ Build

build: ## Build the binary for the current platform
	@mkdir -p $(OUTPUT_DIR)
	go build -buildvcs=false -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY) .
	@echo "  → $(OUTPUT_DIR)/$(BINARY)"

install: ## Install the binary into $$GOPATH/bin
	go install -buildvcs=false -ldflags "$(LDFLAGS)" .

build-all: ## Cross-compile for all supported platforms
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS=$(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH=$(word 2,$(subst /, ,$(platform)))) \
		$(eval EXT=$(if $(filter windows,$(OS)),.exe,)) \
		GOOS=$(OS) GOARCH=$(ARCH) go build \
			-buildvcs=false \
			-ldflags "$(LDFLAGS)" \
			-o $(DIST_DIR)/$(BINARY)-$(OS)-$(ARCH)$(EXT) . ; \
		echo "  → $(DIST_DIR)/$(BINARY)-$(OS)-$(ARCH)$(EXT)" ; \
	)

##@ Testing

test: ## Run the unit test suite
	go test ./... -count=1

test-race: ## Run tests with the race detector
	go test -race ./... -count=1

test-integration: ## Run integration tests (requires Node.js)
	go test -tags=integration ./... -count=1

##@ Code quality

lint: ## Run golangci-lint (install: https://golangci-lint.run/usage/install/)
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go source files
	gofmt -w -s .

##@ Dependencies

deps: ## Download Go module dependencies
	go mod download

tidy: ## Tidy go.mod / go.sum
	go mod tidy

##@ Housekeeping

clean: ## Remove build artefacts
	rm -rf $(OUTPUT_DIR) $(DIST_DIR)

##@ Release

release: ## Run GoReleaser (full release — requires tag and GITHUB_TOKEN)
	goreleaser release --clean

snapshot: ## Run GoReleaser in snapshot mode (local build only)
	goreleaser release --snapshot --clean

##@ Help

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
	  /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help