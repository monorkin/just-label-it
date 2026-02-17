BINARY  := jli
MODULE  := github.com/monorkin/just-label-it
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

GOFLAGS := -trimpath
LDFLAGS := -s -w -X github.com/monorkin/just-label-it/cmd.Version=$(VERSION)

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

# --- Default ---

.PHONY: build
build: ## Build for the current platform into ./bin
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY) .

# --- Multi-platform ---

.PHONY: build-all
build-all: $(PLATFORMS) ## Cross-compile for all platforms into ./dist

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	$(eval GOOS := $(word 1,$(subst /, ,$@)))
	$(eval GOARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS) -ldflags '$(LDFLAGS)' \
		-o dist/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT) .

# --- Release ---

.PHONY: release
release: ## Create a GitHub release (requires TAG, e.g. make release TAG=v0.1.0)
	@if [ -z "$(TAG)" ]; then echo "Error: TAG is required. Usage: make release TAG=v0.1.0" >&2; exit 1; fi
	$(MAKE) build-all VERSION=$(TAG)
	@echo "Creating release $(TAG)..."
	gh release create $(TAG) dist/* \
		--title "$(TAG)" \
		--generate-notes
	@echo "Released $(TAG)"

# --- Housekeeping ---

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin dist

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
