.PHONY: help test test-verbose test-coverage build lint fmt imports vet tidy clean

help: ## Display this help message
	@echo "ZipTax Go SDK - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

test: ## Run tests
	go test ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out | grep total

build: ## Build the SDK
	go build ./...

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code with gofmt
	gofmt -s -w .

imports: ## Organize imports with goimports
	@command -v goimports >/dev/null 2>&1 || { echo "Installing goimports..."; go install golang.org/x/tools/cmd/goimports@latest; }
	goimports -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod and go.sum
	go mod tidy

clean: ## Clean build artifacts and test cache
	go clean -testcache
	rm -f coverage.out coverage.html

check: fmt imports vet lint test ## Run all checks (fmt, imports, vet, lint, test)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed successfully"
