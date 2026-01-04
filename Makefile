BUILD_COMMIT := $(shell git rev-parse HEAD)
BUILD_DIRTY := $(if $(shell git status --porcelain),+CHANGES)
BUILD_COMMIT_FLAG := github.com/rasorp/smuggle-cni/internal/version.BuildCommit=$(BUILD_COMMIT)$(BUILD_DIRTY)

BUILD_TIME ?= $(shell TZ=UTC0 git show -s --format=%cd --date=format-local:'%Y-%m-%dT%H:%M:%SZ' HEAD)
BUILD_TIME_FLAG := github.com/rasorp/smuggle-cni/internal/version.BuildTime=$(BUILD_TIME)

# Populate the ldflags using the Git commit information and and build time
# which will be present in the binary version output.
GO_LDFLAGS = -X $(BUILD_COMMIT_FLAG) -X $(BUILD_TIME_FLAG)

# Disable CGO which is not required for Smuggle CNI.
CGO_ENABLED = 0

bin/%/smuggle-cni: GO_OUT ?= $@
bin/%/smuggle-cni: ## Build Smuggle CNI for GOOS & GOARCH; eg. bin/linux_amd64/smuggle-cni
	@echo "==> Building $@..."
	@GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build \
		-o $(GO_OUT) \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		main.go
	@echo "==> Done"

.PHONY: build
build: ## Build a development version of Smuggle CNI
	@echo "==> Building Smuggle CNI..."
	@go build \
		-o ./bin/smuggle-cni \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		main.go
	@echo "==> Done"

.PHONY: lint
lint: ## Run linters against the Smuggle CNI codebase
	@echo "==> Linting Smuggle CNI..."
	@golangci-lint run --config=build/lint/golangci.yaml ./...
	@echo "==> Done"

HELP_FORMAT="    \033[36m%-22s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
