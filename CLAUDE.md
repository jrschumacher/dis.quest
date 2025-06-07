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

# Git hooks setup (one-time)
lefthook install       # Install git hooks for automatic quality checks
```

### Testing and Quality
```bash
# Run tests
go test ./...          # Direct Go test command
task test             # Via Taskfile
make test             # Via Makefile

# Getting started
task welcome          # Welcome message for new contributors
task quickstart       # Interactive quickstart guide
task help             # Comprehensive help (opens in pager)

# Development workflow shortcuts
task dev-check        # Complete QA check (generate + lint + test)
task quick-fix        # Fast iteration (generate + lint only)
task commit-check     # Pre-commit validation
task dev-setup        # Complete environment setup
task dev-reset        # Reset environment to clean state

# Repository management
task branch-status    # Show comprehensive repo status
task git-cleanup      # Clean up stale branches and worktrees

# Traditional commands (still available)
make lint             # Run golangci-lint v2 with comprehensive checks
make format           # Format with goimports

# GitHub issue management (via Task)
task issue-list       # List open issues
task issue-create     # Create new issue
gh issue view <num>   # View specific issue
gh issue close <num>  # Close completed issue

# Development workflow with worktrees (via Task)
task worktree-dev ISSUE=<num>     # Create worktree from GitHub issue number
task worktree-list                # List all worktrees
task worktree-cleanup BRANCH=<name>  # Clean up worktree and branch after PR merge

# Pull request management (via Task)
task pr-create        # Create PR from current branch
task pr-list          # List open PRs
gh pr view <num>      # View specific PR
gh pr merge <num>     # Merge PR (consider using --squash or --rebase)
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

### Middleware Architecture
- **Chain-Based System**: Clean composition using `middleware.NewChain()`
- **Predefined Chains**: `PublicChain`, `AuthenticatedChain`, `ProtectedChain`
- **Helper Functions**: `WithAuth()`, `WithProtection()`, `WithUserContext()`
- **Custom Chains**: Build specific middleware combinations for different route groups

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
- **Database Engines**: 
  - **SQLite**: Recommended for development/testing (auto-detected from file paths)
  - **PostgreSQL**: Recommended for production (auto-detected from connection strings)
- **Driver Detection**: Automatic based on `database_url` format in config
- **Query Organization**: All SQL queries should be kept in a single query file

### Database Workflow
1. Write SQL queries in the central query file
2. Run `sqlc generate` to generate Go code from SQL
3. Use Tern for database migrations: `tern migrate`
4. Choose database engine by setting `database_url` in config.yaml:
   - SQLite: `./disquest.db` or `file:path/to/db.sqlite`
   - PostgreSQL: `postgres://user:pass@host:port/dbname`

## Development Workflow

### First-Time Setup
For new contributors or clean environments:
```bash
task dev-setup        # Complete environment setup (tools + hooks + generation)
```
This replaces manual tool installation, git hooks setup, and initial code generation.

### Issue-Based Development
1. **Create/Select Issue**: Use `task issue-create` or `task issue-list` to manage work
2. **Create Worktree**: Use `task worktree-dev ISSUE=<number>` to create a dedicated worktree
3. **Develop**: Work in the isolated worktree environment (created in `../dis.quest-issue-<num>/`)
4. **Link Commits**: Use `fixes #<issue-number>` or `closes #<issue-number>` in commit messages
5. **Create PR**: Use `task pr-create` when ready to merge back to main
6. **Post-Merge Cleanup**: Use `task worktree-cleanup BRANCH=issue-<number>` to clean up worktree and branches automatically

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
1. **Start with Issue**: Use `task worktree-dev ISSUE=<number>` for isolated development
2. Make your changes (code, templates, SQL, etc.)
3. **Quality Check**: Use `task dev-check` (replaces steps 3-5 below)
   - Alternative: `task quick-fix` for faster iteration without tests
4. Commit with conventional commit message (use `task commit-check` for validation)
5. **Create PR**: Use `task pr-create` to submit for review
6. **Cleanup**: After PR merge, use `task worktree-cleanup BRANCH=issue-<number>`

### Legacy Development Cycle (manual steps)
1. Run any required code generation (`templ generate`, `sqlc generate`)
2. **MANDATORY: Run `golangci-lint run` and fix all issues**
3. Run tests: `go test ./...`

### Code Quality Requirements
- **Conventional Commits**: Use simple format without scopes: `type: description`
  - Types: `feat`, `fix`, `chore`, `docs`, `style`, `refactor`, `test`, `perf`, `ci`, `build`, `revert`
  - Examples: `feat: add user authentication`, `fix: resolve database connection issue`, `chore: update dependencies`
- **Automated Quality Checks**: Lefthook git hooks handle most quality checks automatically
- Manual checks (if not using lefthook): Run `goimports`, `templ generate`, `sqlc generate`, and `golangci-lint run`
- **CRITICAL**: Always run `golangci-lint run` after ANY unit of work (feature, bug fix, refactor)
- Always test before submitting: `go test ./...`

### Git Hooks (Lefthook)
Pre-commit hooks automatically:
- Format Go code with goimports
- Generate templ and SQLC code when files change
- Run golangci-lint with auto-fix
- Check for TODO/FIXME comments

Pre-push hooks automatically:
- Run full test suite
- Verify build works
- Ensure generated files are up to date

Commit message validation:
- Enforces simple conventional commit format: `type: description`
- No scopes required, keeps commits simple and consistent

### Generated Files (Must Be Committed)
- `*_templ.go` files generated from Templ templates
- Go code generated by SQLC from SQL queries

### OAuth Testing
- Requires valid Bluesky OAuth client configuration
- Test endpoints available at `/auth/*` paths
- Health checks at `/health/*` for service monitoring

## Quick Reference

### Most Common Tasks
```bash
# First time setup
task dev-setup

# Daily development
task worktree-dev ISSUE=123    # Start new feature
task dev-check                 # Ready to commit?
task pr-create                 # Submit for review
task worktree-cleanup BRANCH=issue-123  # After merge

# Troubleshooting
task branch-status             # What's my status?
task git-cleanup               # Clean up repo
task dev-reset                 # Reset environment
```

### Emergency Commands
```bash
task quick-fix                 # Fast check (no tests)
task commit-check              # Validate before commit
task check-tools               # Verify tool installation
```

### Claude Code Assistance
To help Claude Code provide better assistance, use these commands:
```bash
task claude-context           # Full project context for Claude Code
task claude-summary           # Concise project overview
task project-health           # Comprehensive health check
task docs-status              # Documentation freshness check
task help-claude              # Show all Claude Code assistance commands
```

**Recommended:** Run `task claude-context` at the start of Claude Code sessions to provide complete project context.

## Working with Claude Code

### Starting a Session
For optimal Claude Code assistance, provide context at the beginning of each session:
```bash
task claude-context    # Gives Claude Code complete project awareness
```

### During Development
When Claude Code asks about project status or needs updates:
```bash
task project-health    # Comprehensive health check
task docs-status       # Check if documentation needs updating
task branch-status     # Current git and worktree status
```

### Before Making Changes
Help Claude Code understand current state:
```bash
task claude-summary    # Concise overview of current project state
task issue-list        # Show current work items
```

### Optimizing Collaboration
- **Use `task` commands instead of manual git/tool commands** - More efficient, reduces token usage
- **Run `task docs-status`** when asked about documentation updates
- **Use `task help-claude`** to remind yourself of available assistance commands
- **Keep CLAUDE.md updated** as the single source of truth for project workflow
- **See `HUMAN.md`** for advanced prompting techniques and hashtag shortcuts for Claude Code

## Testing Strategy

- **Unit Tests**: Found in `*_test.go` files throughout codebase
- **Key Test Areas**: Authentication (`internal/auth/`), PDS integration (`internal/pds/`), config validation
- **CI/CD**: Automated testing via GitHub Actions on push/PR to main branch
- **Test Isolation**: Tests use Go's standard testing package with table-driven tests where appropriate

## Architectural Decision Records (ADRs)

- ADRs (architectural decision records) are at `/adr`

## Project Documentation

- PRDs (product briefs) are at <root>/docs/prd