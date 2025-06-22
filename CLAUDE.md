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
# Development (requires ngrok for OAuth)
task dev               # Hot reloading with air (recommended)
task watch             # Alias for task dev

# Production/simple
go run . start          # Using Cobra CLI command
task run               # Using Taskfile

# Prerequisites for development
ngrok http 3000        # Required for OAuth (run in separate terminal)
docker-compose up -d   # Start PostgreSQL database
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
task git-cleanup      # Clean up stale branches

# GitHub issue management (via Task)
task issue-list       # List open issues
task issue-create     # Create new issue
gh issue view <num>   # View specific issue
gh issue close <num>  # Close completed issue

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
- Key components: `/pkg/atproto/` for universal logic, `/internal/web/` for HTTP concerns, `server/auth-handlers/` for HTTP layer

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
- **Interactivity**: Datastar for reactive UI components
- **Logging**: Use `internal/logger` for all logging needs

## OAuth + DPoP Implementation Status

### Current State (2025-06-22)
✅ **COMPLETED AND WORKING**: Full end-to-end OAuth with DPoP authentication successfully implemented for ATProtocol/Bluesky integration.

### Custom Lexicon Success (2025-06-22)
🎉 **BREAKTHROUGH ACHIEVED**: Complete custom lexicon implementation working end-to-end with full CRUD operations.

**Proven Working**:
- ✅ Custom lexicon creation: `quest.dis.topic` records successfully stored in production PDS
- ✅ Record retrieval: Individual and collection listing working
- ✅ Test Record: `at://did:plc:qknujfbaxt5ggvbsefz3ixop/quest.dis.topic/topic-1750563554124865000`
- ✅ Key Discovery: `validate: false` required for custom lexicons (learned from WhiteWind analysis)
- ✅ **PDS Browser Interface**: Working create topic modal with form validation and real-time updates
- ✅ **Datastar Integration**: Real-time UI updates with proper signal isolation and fragment merging

### Key Components Working
- **OAuth Provider Abstraction**: Clean interface supporting multiple OAuth implementations (`/internal/oauth/`)
- **Tangled Provider**: Production-ready OAuth using `tangled.sh/icyphox.sh/atproto-oauth` library
- **DPoP Implementation**: RFC-compliant DPoP JWTs with proper nonce handling and access token binding
- **PDS Operations**: Successfully creating and retrieving records from Personal Data Servers
- **Session Management**: Secure DPoP key storage and automatic nonce retry logic

### Implementation Architecture
- **Provider Interface**: `/internal/oauth/provider.go` - Abstraction for OAuth implementations
- **Factory Pattern**: `/internal/oauth/factory.go` - Runtime provider selection
- **Tangled Provider**: `/internal/oauth/tangled.go` - Working OAuth implementation with `private_key_jwt`
- **Manual Provider**: `/internal/oauth/manual.go` - Original implementation preserved for fallback
- **Configuration**: `oauth_provider: tangled` in `config.yaml` for provider selection

### Critical Fixes Applied
- **Client Authentication**: Proper `private_key_jwt` using tangled.sh OAuth library
- **Authorization Headers**: Changed from "Bearer" to "DPoP" per RFC requirements
- **PDS Resolution**: Automatic DID resolution to find user's actual PDS endpoint
- **DPoP Nonce Handling**: Server nonce retry pattern for 401 responses
- **Access Token Binding**: Added `ath` claim to DPoP JWT for token security
- **PAR Integration**: Pushed Authorization Request for DPoP nonce acquisition and auth server discovery
- **Request Body Consumption**: Fixed double request body reading in PDS test handlers
- **Token Refresh DPoP Preservation**: Fixed automatic token refresh to preserve DPoP keys in cookies

### Testing and Validation
- **Dev Interface**: Complete testing interface at `/dev/pds` with comprehensive OAuth and DPoP testing
- **Production Testing**: Successfully creating and retrieving records from real Personal Data Servers
- **Error Handling**: Robust retry logic for DPoP nonce requirements and token expiration
- **Rollback Strategy**: Instant fallback to manual provider via configuration change

### Documentation
- **Implementation Guide**: `/docs/OAUTH_DPOP_IMPLEMENTATION.md` - Complete technical documentation
- **Provider Interface**: Well-documented abstraction allowing future OAuth implementations
- **Security Model**: RFC-compliant DPoP implementation with proper token binding

### File Locations
- **OAuth Interface**: `/internal/oauth/provider.go` - OAuth provider abstraction
- **Tangled Provider**: `/internal/oauth/tangled.go` - Production OAuth implementation  
- **Manual Provider**: `/internal/oauth/manual.go` - Original implementation (preserved)
- **XRPC Client**: `/internal/pds/xrpc.go` - DPoP-enabled PDS operations
- **Session Management**: `/internal/auth/session.go` - DPoP key and nonce handling
- **Dev Interface**: `/server/app/dev.go` - Testing and debugging tools
- **Documentation**: `/docs/OAUTH_DPOP_IMPLEMENTATION.md` - Complete technical guide

### Common Debugging Patterns

#### Request Body Consumption Issues
**Problem**: HTTP request bodies can only be read once. Reading the body in middleware or early handlers prevents later functions from accessing the data.

**Solution**: Parse request data once and pass the parsed data to subsequent functions instead of re-reading the body.

**Example Fix**: Modified `DevPDSTestHandler` to store `parsedData` and pass it to `createTopicFromModal()` instead of having each function read `req.Body`.

#### Token Refresh and Session State
**Problem**: Automatic token refresh must preserve all session state, not just access/refresh tokens. DPoP keys are critical for ATProtocol operations.

**Solution**: When refreshing tokens, explicitly preserve the DPoP key cookie using the `TokenResult.DPoPKey` from the OAuth provider.

**Critical Fix**: 
```go
// In token refresh middleware
if tokenResult.DPoPKey != nil {
    if err := auth.SetDPoPKeyCookie(w, tokenResult.DPoPKey, false); err != nil {
        logger.Error("Failed to set DPoP key cookie after refresh", "error", err)
    }
}
```

### Datastar Integration

#### Version and Setup
- **Current Version**: v1.0.0-beta.11 (Go SDK and JavaScript CDN)
- **CDN**: `https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.0-beta.11/bundles/datastar.js`
- **Go Import**: `github.com/starfederation/datastar/sdk/go`

#### Key Documentation Resources
- **Examples Repository**: https://github.com/starfederation/datastar/tree/main/site
- **Reference Guide**: https://data-star.dev/reference/overview#attribute-plugins
- **Best Practices**: Study the bulk update and edit row examples for component isolation patterns

#### Critical Datastar Patterns

##### Signal Isolation for Multiple Components
**Problem**: Datastar signals are globally scoped by default. Multiple components with the same signal names will interfere with each other.

**Solution**: Use unique signal names with component IDs:
```go
// In Go handler - build unique signal state
initialSignals := make(map[string]any)
for _, msg := range messages {
    initialSignals["liked_"+msg.Id] = msg.Liked
}

// In templ component - use unique signal names
data-on-click={ "$liked_" + signals.Id + " = !$liked_" + signals.Id + "; @post(`/api/messages/" + signals.Id + "/like`)" }
data-show={ "$liked_" + signals.Id }
```

##### Correct Method Syntax
- **Correct**: `@post`, `@get`, `@patch`, `@delete` 
- **Incorrect**: `$$post`, `$$get` (old syntax)
- **Example**: `@post(\`/api/messages/${$messageId}/like\`)`

##### Component State Management
- Use `data-signals` on parent container to initialize global state
- Use `data-bind` for two-way data binding on form elements
- Use `data-show` and `data-hide` for conditional rendering
- Use `data-on-click`, `data-on-change` for event handling

##### Template Literal Interpolation
- Use backticks for dynamic URLs: `@post(\`/api/messages/${$id}/like\`)`
- Avoid string concatenation in datastar expressions when possible

## Configuration

### Required Setup
1. Copy `config.yaml.example` to `config.yaml`
2. Start PostgreSQL: `docker-compose up -d postgres`
3. Run migrations: `task db-migrate`
4. Configure OAuth client settings for ngrok URL (for development)
5. Set up JWKS (JSON Web Key Set) for token signing

### Key Configuration Areas
- `app_env`: Set to "development" or "production"
- `pds_endpoint`: Your Personal Data Server URL
- `oauth_client_id` and `oauth_redirect_url`: Must use ngrok URLs for development OAuth
- `jwks_private`/`jwks_public`: Cryptographic keys for tokens
- `database_url`: PostgreSQL connection string (development via Docker Compose)

### Development OAuth Setup
For local development, OAuth requires publicly accessible URLs:
1. Start ngrok: `ngrok http 3000`
2. Update config.yaml with ngrok URLs:
   - `oauth_client_id`: `https://your-ngrok-url.ngrok.app/auth/client-metadata.json`
   - `oauth_redirect_url`: `https://your-ngrok-url.ngrok.app/auth/callback`
   - `public_domain`: `https://your-ngrok-url.ngrok.app`

## ATProtocol Refactoring (2025-06-22)

### Completed Package Reorganization
✅ **MAJOR ARCHITECTURAL IMPROVEMENT**: Successfully completed comprehensive refactoring to properly separate universal ATProtocol functionality from application-specific code.

### Key Changes Implemented
1. **JWT Utilities Migration**: Migrated `/internal/jwtutil/` → `/pkg/atproto/jwt/` 
   - Eliminated duplicate JWT parsing functionality
   - Consolidated all JWT operations in reusable package

2. **Authentication Simplification**: Eliminated 80% of `/internal/auth/` package
   - Moved universal auth logic to `/pkg/atproto/oauth/` 
   - Moved HTTP-specific concerns to `/internal/web/` (cookies, sessions)
   - Removed unnecessary delegation and abstraction layers

3. **PDS Operations**: Migrated `/internal/pds/` → `/pkg/atproto/pds/`
   - Created lexicon-agnostic PDS client for any ATProtocol application
   - Moved quest.dis.* specific definitions to `/internal/lexicons/`
   - Maintained backward compatibility through LegacyPDSService wrapper

### New Package Structure
- **`/pkg/atproto/`**: Universal ATProtocol functionality (JWT, OAuth, PDS, XRPC)
  - Reusable by any Go application implementing ATProtocol
  - Clean, well-defined APIs following ATProtocol specifications
- **`/internal/web/`**: HTTP-specific web session management  
  - Cookie handling with environment-specific security
  - Session data bridging between HTTP and ATProtocol
- **`/internal/lexicons/`**: Application-specific quest.dis.* lexicon definitions
  - TopicRecord, MessageRecord, ParticipationRecord types
  - Service layer for CRUD operations using generic PDS client

### Benefits Achieved
- **Code Reduction**: Eliminated ~80% of unnecessary abstraction and delegation
- **Clear Separation**: Universal vs application-specific concerns properly separated
- **Reusability**: `/pkg/atproto/` can be used by other Go applications
- **Maintainability**: Simpler, more direct code paths
- **Architectural Clarity**: Proper layering following Go package conventions

### Migration Notes
- All handlers now import `/pkg/atproto/` directly instead of going through `/internal/auth/`
- HTTP cookie management isolated to `/internal/web/` package
- Legacy compatibility maintained through wrapper services during transition
- All compilation errors resolved and build successful

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
- **Query Generation**: SQLC for type-safe Go code generation from SQL (PostgreSQL engine)
- **Migrations**: Tern for database schema migrations
- **Database Engine**: PostgreSQL 17 (via Docker Compose for development)
- **Local Development**: Docker Compose with PostgreSQL 17
- **Production**: PostgreSQL (via Neon or other hosted PostgreSQL)
- **Query Organization**: All SQL queries in `internal/db/queries.sql` with PostgreSQL syntax

### Database Workflow
1. Start PostgreSQL: `docker-compose up -d postgres`
2. Write SQL queries in `internal/db/queries.sql` (PostgreSQL syntax with $1, $2, etc.)
3. Run `sqlc generate` to generate Go code from SQL
4. Create migrations in `migrations/` directory
5. Apply migrations: `task db-migrate`
6. Check status: `task db-status`

### Database Commands
```bash
task db-migrate        # Apply pending migrations
task db-status         # Check migration status
task db-rollback       # Rollback last migration
task db-reset          # Reset and reapply all migrations
```

## Development Workflow

### First-Time Setup
For new contributors or clean environments:
```bash
task dev-setup        # Complete environment setup (tools + hooks + generation)
```
This replaces manual tool installation, git hooks setup, and initial code generation.

### Standard Development Workflow
1. **Create/Select Issue**: Use `task issue-create` or `task issue-list` to manage work
2. **Create Branch**: `git checkout -b feature/issue-description`
3. **Develop**: Make changes, commit with `fixes #<issue-number>` in commit messages
4. **Quality Check**: Run `task dev-check` before committing
5. **Create PR**: Use `task pr-create` when ready to merge back to main

### Template Development
1. Edit `.templ` files in `components/`
2. Run `templ generate` to create corresponding `_templ.go` files
3. **Run `task lint` to check for issues**
4. Build/run normally with Go commands

### Database Development
1. Add SQL queries to `internal/db/queries.sql` (PostgreSQL syntax)
2. Run `sqlc generate` to generate type-safe Go code
3. Create migrations in `migrations/` directory
4. Apply migrations: `task db-migrate`
5. **Run `task lint` to validate generated code**

### Daily Development Cycle
1. Make your changes (code, templates, SQL, etc.)
2. **Quality Check**: Use `task dev-check` for full validation
   - Alternative: `task quick-fix` for faster iteration without tests
3. Commit with conventional commit message
4. **Create PR**: Use `task pr-create` to submit for review

### Code Quality Requirements
- **Conventional Commits**: Use simple format without scopes: `type: description`
  - Types: `feat`, `fix`, `chore`, `docs`, `style`, `refactor`, `test`, `perf`, `ci`, `build`, `revert`
  - Examples: `feat: add user authentication`, `fix: resolve database connection issue`, `chore: update dependencies`
- **Automated Quality Checks**: Lefthook git hooks handle most quality checks automatically
- Manual checks (if not using lefthook): Run `task dev-check` or individual commands like `task lint`
- **CRITICAL**: Always run `task lint` after ANY unit of work (feature, bug fix, refactor)
- Always test before submitting: `task test`

### Git Hooks (Lefthook)
Pre-commit hooks automatically:
- Format Go code with goimports
- Generate templ and SQLC code when files change
- Run linting with auto-fix
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
task dev                       # Start development server with hot reload
task dev-check                 # Ready to commit?
task pr-create                 # Submit for review

# Database
task db-migrate                # Apply migrations
task db-status                 # Check migration status

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
task branch-status     # Current git status
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

## Memories and Guidance

- Always handle go errors. We should always log the error unless its a typical error that will become a hotpoint which will impact the performace of the server. Additionally, when creating errors in Go its preferrable to define them as named vars (e.g. `ErrSomethingHappened`) so that calling libraries and systems and tests can evaluate them rather than comparing the contents of the error string.