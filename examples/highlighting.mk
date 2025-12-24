# Test Makefile for Multi-Language Syntax Highlighting Feature
# This demonstrates automatic language detection and manual overrides

## Build the Go application
build:
	go build -o bin/lazymake ./cmd/lazymake
	go test ./...

## Run Python tests and linting
# language: python
python-demo:
	#!/usr/bin/env python3
	import sys
	def greet(name):
		print(f"Hello, {name}!")
		return True

	if __name__ == "__main__":
		greet("World")
		sys.exit(0)

## Build and run Docker container (shell commands)
docker-demo:
	docker build -t lazymake:test .
	docker run --rm lazymake:test
	docker images | grep lazymake

## Generate Dockerfile
# language: docker
dockerfile-demo:
	cat > Dockerfile << 'EOF'
	FROM golang:1.21-alpine
	WORKDIR /app
	COPY . .
	RUN go build -o /app/lazymake ./cmd/lazymake
	CMD ["/app/lazymake"]
	EOF

## Install and build JavaScript project
npm-demo:
	npm install
	npm run build
	npm test

## Build Rust application
rust-demo:
	cargo build --release
	cargo test
	cargo clippy

## Compile C program
c-demo:
	gcc -o output/app src/main.c
	./output/app

## Compile C++ program
cpp-demo:
	g++ -std=c++17 -o output/app src/main.cpp
	./output/app

## Run Ruby script
ruby-demo:
	ruby script.rb
	bundle install
	bundle exec rspec

## Regular bash commands (auto-detected)
bash-demo:
	echo "Starting build process..."
	for i in {1..5}; do
		echo "Processing item $$i"
	done
	curl -s https://api.github.com/repos/rshelekhov/lazymake | jq '.stars'

## Shell script with shebang
shebang-demo:
	#!/bin/bash
	set -e
	echo "Running with bash shebang"
	ls -la | grep "\.go$$"
	find . -name "*.go" -type f

## Kubernetes deployment
# lang: yaml
k8s-demo:
	kubectl apply -f deploy/k8s/
	kubectl get pods
	helm install myapp ./charts/myapp

## Complex multi-command target
complex-demo:
	@echo "==> Building project..."
	go mod tidy
	go build -ldflags="-s -w" -o bin/app ./cmd/app
	@echo "==> Running tests..."
	go test -v -race ./...
	@echo "==> Build complete!"

## Java compilation and execution
java-demo:
	javac src/Main.java
	java -cp src Main
	mvn clean package

## PHP application
php-demo:
	php artisan migrate
	composer install
	php artisan serve

## Meta target with no recipe
.PHONY: all
all: build test

## Clean build artifacts
clean:
	rm -rf bin/
	rm -rf dist/
	go clean -cache
	find . -name "*.test" -delete

# =============================================================================
# Full-Stack Web Application Build Pipeline (for demo screenshots)
# Demonstrates: dependency graph, variables, syntax highlighting, parallel execution
# =============================================================================

# Build configuration variables
VERSION := 1.2.3
BUILD_DIR := ./dist
FRONTEND_DIR := ./web/frontend
BACKEND_DIR := ./cmd/server
DOCKER_REGISTRY := registry.example.com
export NODE_ENV := production
export CGO_ENABLED := 0

## Install all project dependencies
install-deps:
	@echo "Installing dependencies..."
	npm install --prefix $(FRONTEND_DIR)
	go mod download
	go mod verify

## Build React frontend application
# language: javascript
build-frontend: install-deps
	@echo "Building frontend v$(VERSION)..."
	cd $(FRONTEND_DIR) && npm run build
	mkdir -p $(BUILD_DIR)/static
	cp -r $(FRONTEND_DIR)/build/* $(BUILD_DIR)/static/

## Build Go backend server
build-backend: install-deps
	@echo "Building backend v$(VERSION)..."
	go build -ldflags="-s -w -X main.Version=$(VERSION)" \
		-o $(BUILD_DIR)/server $(BACKEND_DIR)
	chmod +x $(BUILD_DIR)/server

## Run frontend tests with coverage
test-frontend: build-frontend
	cd $(FRONTEND_DIR) && npm run test -- --coverage --watchAll=false
	cd $(FRONTEND_DIR) && npm run lint

## Run backend tests and benchmarks
test-backend: build-backend
	go test -v -race -coverprofile=coverage.out ./...
	go test -bench=. -benchmem ./internal/...
	go vet ./...

## Generate and optimize Docker image
# language: docker
build-docker: test-frontend test-backend
	cat > $(BUILD_DIR)/Dockerfile << 'EOF'
	FROM alpine:3.18
	RUN apk add --no-cache ca-certificates tzdata
	WORKDIR /app
	COPY server /app/
	COPY static /app/static/
	EXPOSE 8080
	USER nobody
	CMD ["/app/server"]
	EOF
	docker build -t $(DOCKER_REGISTRY)/myapp:$(VERSION) $(BUILD_DIR)
	docker tag $(DOCKER_REGISTRY)/myapp:$(VERSION) $(DOCKER_REGISTRY)/myapp:latest

## Run integration tests
# language: python
test-integration: build-docker
	#!/usr/bin/env python3
	import requests
	import time
	import subprocess

	# Start container
	container = subprocess.Popen(
		["docker", "run", "-p", "8080:8080", "--rm",
		 f"$(DOCKER_REGISTRY)/myapp:$(VERSION)"],
		stdout=subprocess.PIPE
	)

	time.sleep(3)  # Wait for startup

	# Run integration tests
	response = requests.get("http://localhost:8080/health")
	assert response.status_code == 200
	assert response.json()["version"] == "$(VERSION)"

	container.terminate()
	print("✓ Integration tests passed")

## Deploy to Kubernetes cluster
# language: yaml
deploy-staging: test-integration
	@echo "Deploying version $(VERSION) to staging..."
	kubectl set image deployment/myapp \
		myapp=$(DOCKER_REGISTRY)/myapp:$(VERSION) \
		--namespace=staging
	kubectl rollout status deployment/myapp --namespace=staging
	kubectl get pods --namespace=staging -l app=myapp

## Full production deployment with health checks
deploy-production: test-integration
	@echo "Deploying version $(VERSION) to production..."
	helm upgrade --install myapp ./charts/myapp \
		--set image.tag=$(VERSION) \
		--set image.repository=$(DOCKER_REGISTRY)/myapp \
		--set replicas=3 \
		--namespace=production \
		--wait --timeout=5m
	kubectl get pods --namespace=production -l app=myapp
	@echo "✓ Deployment complete: v$(VERSION)"

## Database migration
# language: sql
migrate-db: deploy-staging
	@echo "Running database migrations..."
	psql $(DATABASE_URL) << 'EOF'
	-- Create users table
	CREATE TABLE IF NOT EXISTS users (
	    id SERIAL PRIMARY KEY,
	    email VARCHAR(255) UNIQUE NOT NULL,
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Add indexes
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_created ON users(created_at);
	EOF

## Complete CI/CD pipeline (meta target)
.PHONY: ci-pipeline
ci-pipeline: build-frontend build-backend test-frontend test-backend build-docker test-integration deploy-staging
