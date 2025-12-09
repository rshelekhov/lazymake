.PHONY: build install test clean run release snapshot

build: ## Build the application
	go build -o lazymake cmd/lazymake/main.go

install: ## Install to GOPATH/bin (requires GOPATH/bin in PATH)
	go install ./cmd/lazymake

install-system: ## Install to /usr/local/bin (requires sudo)
	go build -o lazymake cmd/lazymake/main.go
	sudo mv lazymake /usr/local/bin/

run: ## Run the application without installing
	go run cmd/lazymake/main.go

test: ## Run all tests
	go test ./...

clean: ## Clean build artifacts
	rm -f lazymake

snapshot: ## Test release build locally (doesn't publish)
	goreleaser release --snapshot --clean

release: ## Create a new release (requires git tag)
	@echo "To create a release:"
	@echo "1. Create and push a tag: git tag -a v0.1.0 -m 'Release v0.1.0'"
	@echo "2. Push the tag: git push origin v0.1.0"
	@echo "3. GitHub Actions will automatically build and publish the release"