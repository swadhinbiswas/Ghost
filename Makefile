.DEFAULT_GOAL := help
BINARY := ghost

.PHONY: help build run test test-race vet fmt fmt-check lint tidy clean site-dev site-build ci

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

build: ## Build the ghost binary
	go build -o $(BINARY) ./main.go

run: ## Run from source
	go run ./main.go

test: ## Run tests
	go test ./...

test-race: ## Run tests with the race detector and coverage
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format Go code
	gofmt -w .

fmt-check: ## Check formatting (fails if unformatted)
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files:"; gofmt -l .; exit 1)

lint: ## Run golangci-lint (must be installed)
	golangci-lint run

tidy: ## Tidy go modules
	go mod tidy

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out

site-dev: ## Run the website locally (Bun)
	cd site && bun install && bun run dev

site-build: ## Build the website (Bun)
	cd site && bun install && bun run build

ci: fmt-check vet test ## Run the local CI checks
