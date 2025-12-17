.PHONY: help run build test clean docker-up docker-down migrate-up migrate-down ent-generate air-init dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/server cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

docker-up: ## Start docker containers
	docker-compose up -d

docker-down: ## Stop docker containers
	docker-compose down

docker-logs: ## View docker logs
	docker-compose logs -f

ent-new: ## Create new ent schema (usage: make ent-new name=SchemaName)
	go run -mod=mod entgo.io/ent/cmd/ent new --target internal/repository/schema $(name)

ent-generate: ## Generate ent code
	go generate ./internal/repository

air-init: ## Initialize air configuration
	air init

dev: ## Run with air hot reload
	air

install-tools: ## Install required tools
	go install entgo.io/ent/cmd/ent@latest
	go install github.com/air-verse/air@latest

deps: ## Download dependencies
	go mod download
	go mod tidy

.DEFAULT_GOAL := help
