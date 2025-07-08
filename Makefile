.PHONY: help build clean migrate-up migrate-down migrate-create migrate-version gen deps lint fmt

# Database connection string
DB_URL := "postgres://user:password@localhost:5432/outbox_db?sslmode=disable"

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-15s %s\n", $$1, $$2}'

build: ## Build all binaries to bin/ directory
	go build -o bin/api cmd/api/main.go
	go build -o bin/publisher cmd/publisher/main.go
	go build -o bin/consumer cmd/consumer/main.go

clean: ## Remove built binaries
	rm -f bin/*

migrate-up: ## Run all pending migrations
	migrate -database $(DB_URL) -path db/migrations up

migrate-down: ## Reset all migrations
	migrate -database $(DB_URL) -path db/migrations down -all

migrate-version: ## Show current migration version
	migrate -database $(DB_URL) -path db/migrations version

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir db/migrations -seq $(NAME)

gen: ## Generate Go code from SQL
	sqlc generate --file db/sqlc.yaml

deps: ## Install dependencies and tools
	go mod tidy
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code with golangci-lint
	golangci-lint fmt
