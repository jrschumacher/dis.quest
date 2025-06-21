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
- **Interactivity**: Datastar for reactive UI components
- **Logging**: Use `internal/logger` for all logging needs

## DPoP Implementation Status

### Current State (2025-06-21)
✅ **Complete DPoP Implementation**: Full end-to-end DPoP (Demonstration of Proof of Possession) authentication is implemented and working correctly for ATProtocol OAuth.

### Key Components Working
- **OAuth Flow**: Successfully authenticating with Bluesky using `"atproto transition:generic"` scope
- **DPoP Key Management**: Keys generated, stored in cookies, and retrieved properly
- **DPoP JWT Creation**: `auth.CreateDPoPJWT()` creating valid DPoP JWTs with correct headers
- **XRPC Integration**: Both Authorization Bearer tokens and DPoP headers sent to PDS
- **Session Management**: Access tokens and DPoP keys properly extracted from user sessions

### Implementation Details
- **Scope Format**: Using `[]string{"atproto", "transition:generic"}` (separate array elements work correctly)
- **DPoP Methods**: `CreateRecordWithDPoP()` and `GetRecordWithDPoP()` implemented in XRPC client
- **Dev Interface**: Complete testing interface at `/dev/pds` with Force Re-authenticate functionality
- **Logging**: Comprehensive debugging for OAuth flow, DPoP creation, and PDS requests

### Current Issue: OAuth Client Authentication Method Missing
🚫 **Root Cause Identified**: The "Bad token scope" error is due to missing `private_key_jwt` client authentication.

**Key Discovery (2025-06-21 14:20)**: 
- Our DPoP implementation is **technically correct** - JWK thumbprints match perfectly
- **Tangled.sh analysis** revealed they use `token_endpoint_auth_method: "private_key_jwt"` 
- **WhiteWind analysis** showed they bypass OAuth entirely using session-based auth
- We're using `"none"` authentication which is insufficient for DPoP flows

**Technical Details**:
- ✅ **DPoP JWT Structure**: Correct with fixed IEEE P1363 signature encoding (64 bytes)
- ✅ **JWK Thumbprint Binding**: Perfect match between token `jkt` and calculated thumbprint
- ✅ **Scope Format**: Using `"atproto transition:generic"` (matches Tangled.sh)
- ✅ **Cookie Clearing Fix**: DPoP keys now properly cleared during re-authentication
- ❌ **Client Authentication**: Missing `private_key_jwt` for token exchange

### Critical OAuth Flow Requirements for DPoP
**Problem**: DPoP requires two separate authentication mechanisms:
1. **Client Authentication**: `private_key_jwt` to authenticate the OAuth client during token exchange
2. **Proof of Possession**: DPoP headers to prove possession of the bound key

**Why This Matters**: 
- DPoP keys should remain client-side for security
- Server needs separate client authentication for token requests  
- Traditional IdPs allow pre-registering client keys; ATProtocol requires dynamic client metadata

### Implementation Status
🔄 **Partial Fix Applied**: 
- Updated client metadata to include `private_key_jwt` method and JWKS
- Still need to implement client assertion JWT creation for token exchange
- Complex manual implementation suggests using specialized library

### Next Steps for Resolution
1. **Complete `private_key_jwt` Implementation**: Create client assertion JWTs for token exchange
2. **Alternative: Use ATProtocol OAuth Library**: Follow Tangled.sh's approach with specialized library
3. **Alternative: Session-Based Auth**: Implement WhiteWind's approach to bypass OAuth restrictions

### File Locations
- **DPoP Implementation**: `/internal/pds/xrpc.go` - DPoP JWT headers
- **ATProtocol Service**: `/internal/pds/atproto.go` - Service layer with DPoP support  
- **OAuth Configuration**: `/internal/auth/auth.go` - Scope and OAuth2 config
- **Dev Interface**: `/server/app/dev.go` - Testing and debugging tools
- **Templates**: `/components/dev_pds.templ` - UI for testing DPoP functionality

### Critical Discoveries
- **Client Auth Missing**: `private_key_jwt` required for DPoP flows, not just DPoP headers alone
- **Two-Key Architecture**: Need separate keys for client auth and DPoP proof of possession  
- **Scope Format Critical**: Must use exact string `"atproto transition:generic"` (not array)
- **Signature Encoding**: ES256 requires IEEE P1363 format (fixed 32+32 bytes for P-256)
- **Cookie Clearing**: DPoP keys must be cleared during re-auth to prevent key/token mismatch
- **Library Complexity**: Manual DPoP+OAuth implementation extremely complex vs specialized libraries
- **Error Misleading**: "Bad token scope" actually means authentication method issues, not scopes

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