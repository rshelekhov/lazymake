# Example Makefile demonstrating lazymake's safety features
#
# This Makefile contains various dangerous commands to showcase
# how lazymake detects and warns about potentially destructive operations.
#
# Try it with: lazymake -f examples/dangerous.mk
#
# Expected indicators:
# 🚨 deploy-prod  (CRITICAL - requires confirmation)
# 🚨 nuke-db      (CRITICAL - requires confirmation)
# ⚠️  clean       (WARNING - downgraded from CRITICAL for clean target)
# ⚠️  docker-clean (WARNING)

.PHONY: build test clean deploy-prod nuke-db docker-clean safe-target

## Build the application
build:
	go build -o app ./cmd/app

## Run tests
test:
	go test ./...

## Clean build artifacts
clean:
	rm -rf build/
	rm -f app

## Deploy to production (DANGEROUS!)
deploy-prod:
	kubectl apply -f k8s/prod/
	terraform apply -var-file=prod.tfvars

## Drop production database (CRITICAL!)
nuke-db:
	psql -c 'DROP DATABASE production;'

## Clean Docker resources
docker-clean:
	docker system prune -f
	docker volume prune -f

## Safe target with no recipe
safe-target:
	echo "This is safe"
	echo "Hello, world!"
