# Go SDK for the DBaaS Platform API.

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run unit tests and examples against the in-memory fake API.
	go test -race ./...

.PHONY: test-integration
test-integration: ## Run the integration test against a real API, configured by .env.
	@test -f .env || { echo "sdk/.env is missing: copy .env.example and fill it in"; exit 1; }
	set -a && . ./.env && set +a && go test -race -tags integration -run Integration -count=1 -v ./...

.PHONY: vet
vet: ## Check formatting and run go vet.
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "run: gofmt -w ."; exit 1; }
	go vet ./...
	go vet -tags integration ./...

.PHONY: cover
cover: ## Show test coverage per function.
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: docs
docs: ## Serve the package docs, examples included, at http://localhost:6060.
	go run golang.org/x/pkgsite/cmd/pkgsite@latest -http localhost:6060 .
