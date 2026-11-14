# Go SDK for the DBaaS Platform API.

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

## Pinned tool versions. Tools are downloaded into ./bin, nothing is installed globally.
GOLANGCI_LINT_VERSION ?= v2.15.0

LOCALBIN      ?= $(CURDIR)/bin
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

.PHONY: all
all: help

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: test
test: ## Run unit tests and examples against the in-memory fake API.
	go test -race ./...

.PHONY: test-integration
test-integration: ## Run the integration test against a real API, configured by .env.
	@test -f .env || { echo "sdk/.env is missing: copy .env.example and fill it in"; exit 1; }
	set -a && . ./.env && set +a && go test -race -tags integration -run Integration -count=1 -v ./...

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run golangci-lint (formatting, vet and style checks).
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT) ## Run golangci-lint and apply its fixes.
	$(GOLANGCI_LINT) run --fix

.PHONY: cover
cover: ## Show test coverage per function.
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: docs
docs: ## Serve the package docs, examples included, at http://localhost:6060.
	go run golang.org/x/pkgsite/cmd/pkgsite@latest -http localhost:6060 .

##@ Tools

$(GOLANGCI_LINT):
	@mkdir -p $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	mv $(LOCALBIN)/golangci-lint $@
