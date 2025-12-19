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
