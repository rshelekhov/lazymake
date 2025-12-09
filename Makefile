.PHONY: build install test clean run release snapshot

# Build the application
build:
	go build -o lazymake cmd/lazymake/main.go

# Install to GOPATH/bin (requires GOPATH/bin in PATH)
install:
	go install ./cmd/lazymake

# Install to /usr/local/bin (requires sudo)
install-system:
	go build -o lazymake cmd/lazymake/main.go
	sudo mv lazymake /usr/local/bin/

# Run the application without installing
run:
	go run cmd/lazymake/main.go

# Run all tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f lazymake

# Test release build locally (doesn't publish)
snapshot:
	goreleaser release --snapshot --clean

# Create a new release (requires git tag)
release:
	@echo "To create a release:"
	@echo "1. Create and push a tag: git tag -a v0.1.0 -m 'Release v0.1.0'"
	@echo "2. Push the tag: git push origin v0.1.0"
	@echo "3. GitHub Actions will automatically build and publish the release"