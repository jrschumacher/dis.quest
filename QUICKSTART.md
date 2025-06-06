# dis.quest Development Quickstart

Welcome to dis.quest! This guide will get you up and running quickly.

## 🚀 First-Time Setup

```bash
# Complete environment setup (tools + hooks + generation)
task dev-setup
```

This installs all required tools, sets up git hooks, and generates initial code.

## 🎯 Daily Development Workflow

### Start New Work
```bash
# List available issues
task issue-list

# Create worktree for issue #123
task worktree-dev ISSUE=123

# Switch to the new worktree
cd ../dis.quest-issue-123
```

### Make Changes
```bash
# Quick iteration (generate + lint, no tests)
task quick-fix

# Full quality check (generate + lint + test)
task dev-check

# Pre-commit validation
task commit-check
```

### Submit & Cleanup
```bash
# Create pull request
task pr-create

# After PR merge, cleanup (from main worktree)
cd ../dis.quest
task worktree-cleanup BRANCH=issue-123
```

## 🔧 Essential Commands

### Development
- `task run` - Start the development server
- `task test` - Run tests
- `task lint` - Run linter
- `task dev-check` - Complete quality check
- `task dev-reset` - Reset environment to clean state

### Project Management
- `task issue-list` - Show open GitHub issues
- `task issue-create` - Create new issue
- `task pr-list` - Show open pull requests
- `task branch-status` - Show comprehensive repository status

### Troubleshooting
- `task check-tools` - Verify tool installation
- `task git-cleanup` - Clean up stale branches and worktrees
- `task project-health` - Comprehensive health check

## 🤖 Working with Claude Code

### Start Claude Code Sessions
```bash
# Provide complete project context to Claude Code
task claude-context
```

### During Development
```bash
# Quick project overview
task claude-summary

# Check documentation status
task docs-status

# Show all Claude Code assistance commands
task help-claude
```

### Advanced Prompting
See `HUMAN.md` for:
- Hashtag shortcuts (#explain, #verify, #safe, etc.)
- Project-specific prompting patterns
- Template prompts for common tasks
- Pro tips for effective collaboration

## 📁 Project Structure

```
dis.quest/
├── cmd/                    # CLI commands
├── internal/               # Internal packages
│   ├── auth/              # Authentication logic
│   ├── config/            # Configuration management
│   └── pds/               # Personal Data Server integration
├── server/                # HTTP handlers organized by domain
│   ├── app/               # Main application routes
│   ├── auth-handlers/     # OAuth authentication
│   └── health-handlers/   # Health checks
├── components/            # Templ templates
├── migrations/            # Database migrations
├── CLAUDE.md             # Comprehensive development guide
└── Taskfile.yml          # Task automation
```

## 🏗️ Architecture Overview

- **Backend**: Go with Templ templates for server-side rendering
- **Styling**: Pico CSS (minimal, classless framework)
- **Database**: SQLite with SQLC for type-safe queries
- **Authentication**: OAuth2 with ATProtocol/Bluesky
- **Data**: Custom lexicons under `quest.dis.*` namespace

## 📚 Key Files

- `CLAUDE.md` - Complete development workflow and architecture guide
- `Taskfile.yml` - All development tasks and commands
- `config.yaml.example` - Configuration template
- `lefthook.yml` - Git hooks configuration

## 🆘 Getting Help

```bash
# Show all available tasks
task --list

# Show Claude Code assistance commands
task help-claude

# Show project health and next steps
task project-health

# Show documentation status
task docs-status
```

## 💡 Tips

- Use `task` commands instead of direct git/tool commands for consistency
- Run `task dev-check` before committing changes
- Keep worktrees organized with the `task worktree-*` commands
- Use GitHub issues for all new work
- Refer to `CLAUDE.md` for comprehensive documentation

---

Happy coding! 🎉

For detailed information, see `CLAUDE.md` or run `task help-claude` for Claude Code optimization.