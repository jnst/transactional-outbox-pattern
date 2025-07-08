# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of the Transactional Outbox Pattern using PostgreSQL and Redis Streams. The pattern ensures reliable message delivery by storing outbox events in the same database transaction as business data, then asynchronously publishing them to Redis Streams.

**Requirements**: Go 1.23+ (uses caarlos0/env/v11, pgx/v5, rueidis)

## Architecture

The project follows a 3-component architecture:

1. **API Server** (`cmd/api/`) - Handles user requests and creates outbox events in transactions
2. **Outbox Publisher** (`cmd/publisher/`) - Polls unpublished events and publishes to Redis Streams  
3. **Message Consumer** (`cmd/consumer/`) - Consumes messages from Redis Streams and processes events

Key architectural principles:
- **SOLID principles**: Interfaces and dependency injection are used throughout
- **Repository pattern**: Database operations abstracted behind interfaces
- **Service layer**: Business logic separated from HTTP handlers
- **Clean architecture**: Dependencies point inward (cmd → internal → pkg)

## Code Structure

- `internal/model/` - Domain models and validation
- `internal/repository/` - Data access layer with PostgreSQL implementations
- `internal/service/` - Business logic layer
- `internal/db/` - Generated sqlc code for type-safe SQL operations
- `internal/config/` - Environment configuration management
- `internal/logger/` - Structured logging setup with slog
- `db/migrations/` - Database schema migrations
- `db/queries/` - SQL queries for sqlc generation
- `db/sqlc.yaml` - sqlc configuration file

## Essential Commands

### Development Setup
```bash
make deps          # Install required tools (migrate, sqlc)
make migrate-up    # Run database migrations
make gen          # Generate Go code from SQL queries
make build        # Build all binaries to bin/ directory
```

### Code Quality
```bash
make lint         # Run golangci-lint (configured for v2 with standard preset)
make fmt          # Format code with golangci-lint formatters
```

### Database Operations
```bash
make migrate-create NAME=description  # Create new migration
make migrate-down                     # Reset all migrations
make migrate-version                  # Check current migration version
```

### Running Services
```bash
# Start infrastructure first
docker-compose up -d

# Build binaries (optional - can also use go run)
make build

# Run services in separate terminals
./bin/api        # API Server (request handler)
./bin/publisher  # Outbox Publisher (polling events)  
./bin/consumer   # Message Consumer (processing events)

# Or run directly with go run
go run cmd/api/main.go        # API Server
go run cmd/publisher/main.go  # Outbox Publisher
go run cmd/consumer/main.go   # Message Consumer
```

## Configuration

Environment variables managed via `caarlos0/env/v11` with struct-based configuration:
- `DATABASE_URL` - PostgreSQL connection (default: postgres://user:password@localhost:5432/outbox_db?sslmode=disable)
- `REDIS_ADDR` - Redis address (default: localhost:6379)
- `PORT` - API server port (default: 8080)
- `PUBLISHER_POLL_INTERVAL` - Polling frequency (default: 5s)
- `PUBLISHER_BATCH_SIZE` - Events per batch (default: 10)
- `CONSUMER_NAME` - Consumer identifier (default: consumer-1)
- `LOG_LEVEL` - Logging level (default: info)

Configuration struct in `internal/config/config.go` uses `envDefault` tags for defaults.

## Database Schema

Two main tables:
- `users` - Business data (id, name, email, created_at)
- `outbox_events` - Transactional outbox (id, aggregate_id, event_type, payload, created_at, published_at)

The `published_at` column tracks publication status - NULL means unpublished.

## Important Implementation Details

### Transaction Management
- `TransactionManagerImpl` handles explicit rollback only on errors
- Uses pgx/v5 connection pools for PostgreSQL
- Outbox events created in same transaction as business data
- Repository pattern abstracts database operations behind interfaces

### SQL Code Generation
- Uses sqlc v2 with `query_parameter_limit: 1` to generate parameter structs
- Configuration file located at `db/sqlc.yaml`
- SQL queries in `db/queries/` generate Go code in `internal/db/`
- After modifying SQL queries, run `make gen` to regenerate Go code
- Uses `published_at IS NULL` to identify unpublished events

### Linting Configuration
- golangci-lint v2 with `default: fast` preset
- Key disabled linters: `depguard`, `lll`, `wsl`, `wsl_v5` 
- `revive.exported` rule enabled - requires comments on all exported types/functions
- Formatters: `gofmt`, `golines`, `gci` (import grouping enforced)
- Import ordering: standard → default → project prefix
- Test files excluded from errcheck, gosec, dupl

### Redis Integration
- Uses rueidis client for Redis Streams
- Events published to `user:events` stream with Consumer Groups
- Publisher marks events as published by setting `published_at` timestamp
- Consumer uses `email-service` group for distributed processing
- Stream commands: XADD (publish), XREADGROUP (consume), XACK (acknowledge)

### Logging
- Uses Go's structured logging (slog) with text format for human-readable output
- Centralized logger setup in `internal/logger/` package
- Type-safe field functions: `slog.String()`, `slog.Int64()`, `slog.Duration()`, etc.
- Log levels: debug, info, warn, error (configured via `LOG_LEVEL` environment variable)
- Structured fields include: event_id, user_id, stream, service, error, message_id
- All services (api, publisher, consumer) use consistent logging patterns

## Testing API

```bash
# Create user (triggers outbox event)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Check Redis streams
docker exec -it redis redis-cli XREAD STREAMS user:events 0-0
```