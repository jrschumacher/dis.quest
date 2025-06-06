# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`dis.quest` is an experimental Go web application that implements a secure discussion platform built on ATProtocol (the protocol used by Bluesky). It provides GitHub-style discussions with topics, replies, and participation tracking, with optional OpenTDF encryption for enhanced security.

### Project Philosophy
- Users authenticate with ATProtocol and manage their own data via their Personal Data Server (PDS)
- Each user's data is structured using lexicons and stored in their own PDS whenever possible
- The app aims for lightweight, interoperable experiences within the ATProtocol ecosystem

## Development Commands

### Starting the Server
```bash
# Primary methods
go run main.go          # Using main.go directly
go run . start          # Using Cobra CLI command
task run               # Using Taskfile
make run               # Using Makefile

# Development with hot reloading
task watch             # Uses reflex to watch .go and .templ files
task dev               # Uses air for hot reloading
```

### Building
```bash
# Generate templates and build
task build             # Runs templ generate + go build
make build             # Standard go build to bin/disquest

# Code generation (required before building)
templ generate         # Generates Go code from .templ files
sqlc generate          # Generates Go code from SQL queries
```

### Testing and Quality
```bash
# Run tests
go test ./...          # Direct Go test command
task test             # Via Taskfile
make test             # Via Makefile

# Linting and formatting
make lint             # Run golangci-lint v2 with comprehensive checks
make format           # Format with goimports
```

## Architecture Overview

### Core Structure
- **Entry Point**: `main.go` → `cmd/` (Cobra CLI) → `server/server.go`
- **HTTP Routing**: Domain-organized handlers in `server/` subdirectories
- **Templates**: Templ-based templates in `components/` (must run `templ generate`)
- **Configuration**: Viper-based config loading from `config.yaml`

### Handler Organization
```
server/
├── app/                    # Main application routes ("/")
├── auth-handlers/          # OAuth authentication ("/auth")
├── health-handlers/        # Health checks ("/health")
└── dot-well-known-handlers/ # .well-known endpoints
```

### Authentication Flow
- OAuth2 with Bluesky/ATProtocol
- PKCE and DPoP for enhanced security
- Session management with secure cookies
- Key components: `internal/auth/` for logic, `server/auth-handlers/` for HTTP layer

### Data Model Principles
- Authentication uses ATProtocol and Bluesky identity
- Each user's lexicon data must be stored in their own PDS
- Exception: Public data (e.g., Topic fields) shared across users
- All interactions and data writes are scoped to the authenticated user

### Frontend Architecture
- **Templating**: Templ (type-safe Go templates)
- **Styling**: Pico CSS (minimal, classless CSS framework)
- **Interactivity**: Server-side rendering with minimal JavaScript
- **Static Assets**: Served from `assets/` directory

### Technology Stack
- **Database**: SQLC for type-safe SQL queries
- **HTML Views**: Templ for server-side rendering
- **Styling**: Pico CSS (avoid React or heavy JavaScript frameworks)
- **Logging**: Use `internal/logger` for all logging needs

## Configuration

### Required Setup
1. Copy `config.yaml.example` to `config.yaml`
2. Configure OAuth client settings for Bluesky integration
3. Set up JWKS (JSON Web Key Set) for token signing
4. Configure PDS endpoint and database connection

### Key Configuration Areas
- `app_env`: Set to "development" or "production"
- `pds_endpoint`: Your Personal Data Server URL
- `oauth_client_id` and `oauth_redirect_url`: OAuth integration
- `jwks_private`/`jwks_public`: Cryptographic keys for tokens
- `database_url`: Database connection string (configurable for SQLite or PostgreSQL)

## ATProtocol Integration

### Custom Lexicons
The application defines custom lexicons under `quest.dis.*`:
- `quest.dis.topic`: Discussion topics with metadata
- `quest.dis.message`: Messages within topics
- `quest.dis.participation`: User participation tracking

### Lexicon Files
- Development: `api/disquest/` - working definitions
- Generated: `lexicons/` - final lexicon files

## Database Architecture

### Database Stack
- **Query Generation**: SQLC for type-safe Go code generation from SQL
- **Migrations**: Tern for database schema migrations
- **Database Engine**: SQLite by default, configurable via `database_url` in config
- **Query Organization**: All SQL queries should be kept in a single query file

### Database Workflow
1. Write SQL queries in the central query file
2. Run `sqlc generate` to generate Go code from SQL
3. Use Tern for database migrations: `tern migrate`
4. Database engine is configurable via connection string in config.yaml

## Development Workflow

### Template Development
1. Edit `.templ` files in `components/`
2. Run `templ generate` to create corresponding `_templ.go` files
3. **Run `golangci-lint run` to check for issues**
4. Build/run normally with Go commands

### Database Development
1. Add SQL queries to the central query file
2. Run `sqlc generate` to generate type-safe Go code
3. Create migrations using Tern for schema changes
4. **Run `golangci-lint run` to validate generated code**
5. Test with SQLite locally, deploy with configurable database

### Standard Development Cycle
1. Make your changes (code, templates, SQL, etc.)
2. Run any required code generation (`templ generate`, `sqlc generate`)
3. **MANDATORY: Run `golangci-lint run` and fix all issues**
4. Run tests: `go test ./...`
5. Commit with conventional commit message

### Code Quality Requirements
- Follow Conventional Commits for all commit messages and PR titles
- Run `goimports` before committing (handles both formatting and import management)
- If `.templ` files are changed, run `templ generate`
- **CRITICAL**: Always run `golangci-lint run` after ANY unit of work (feature, bug fix, refactor)
- Always test before submitting: `go test ./...`

### Generated Files (Must Be Committed)
- `*_templ.go` files generated from Templ templates
- Go code generated by SQLC from SQL queries

### OAuth Testing
- Requires valid Bluesky OAuth client configuration
- Test endpoints available at `/auth/*` paths
- Health checks at `/health/*` for service monitoring

## Testing Strategy

- **Unit Tests**: Found in `*_test.go` files throughout codebase
- **Key Test Areas**: Authentication (`internal/auth/`), PDS integration (`internal/pds/`), config validation
- **CI/CD**: Automated testing via GitHub Actions on push/PR to main branch
- **Test Isolation**: Tests use Go's standard testing package with table-driven tests where appropriate