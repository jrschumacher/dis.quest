# 🧭 Dis.quest – Structured Discussions on ATProtocol

## Overview
**Dis.quest** is a structured, topic-based discussion protocol and app view built on top of [ATProtocol](https://atproto.com/). It introduces a GitHub Discussions–style model for decentralized communities, enabling asynchronous conversation around shared topics—scoped by participation, threaded replies, reactions, and optional Q&A flow.

While **OpenTDF** integration is supported through optional lexicons, the base protocol and experience are designed to function independently, with a clear separation between public and protected messaging.

## Goals
Dis.quest aims to define a protocol-native way to:
- Start and organize **topics** scoped around a defined subject
- Support **replies** and **threads** in a structured conversation
- Track **participation** and notify followers of updates
- Enable **optional Q&A**, **reactions**, and **app-level categorization**
- Provide an idiomatic, reference implementation in **Go**

## Lexicon Namespace

| Lexicon ID                     | Purpose                          |
| ------------------------------ | -------------------------------- |
| `quest.dis.topic`              | Defines a discussion topic       |
| `quest.dis.post`               | Contribution post or reply       |
| `quest.dis.participation`      | Follow or moderation signal      |
| `quest.dis.sec.*` *(optional)* | OpenTDF-encrypted message fields |

## Key Features

- Topics with title, description, and categories
- Posts scoped to a topic, supporting threaded replies
- Participation model for follows and moderation
- Q&A topics with `selectedAnswer` support
- Emoji reactions and upvotes
- Mentions (`@handle`) for user notifications *(planned)*
- Attachments with optional encryption
- Defined categories for filtering/browsing
- Idiomatic Go implementation and PDS/AppView compatibility

## Scope

### ✅ In Scope
- Discussion topics and structured replies
- Optional `selectedAnswer` (Q&A-style topics)
- Threaded replies (shallow depth)
- Emoji reactions and upvotes
- Attachment references using ATProto blob store
- Participation and follower tracking
- Predefined categories
- Optional WebSocket-based subscriptions

### 🚫 Out of Scope
- Rich text formatting or WYSIWYG editing
- Direct messaging or group chats
- Private topics or per-message ACLs
- Voting-based features (e.g., polls)

## OpenTDF Integration
OpenTDF integration is optional and scoped under the `quest.dis.sec.*` lexicons. When enabled, messages and attachments can be encrypted with attribute-based access control policies. This integration demonstrates the extensibility of ATProtocol to support secure messaging—but is not required for base app functionality.

## Future Directions
- Moderation roles and admin actions
- Nested threading and reply hierarchies
- Reaction badges and reputation signals
- Notification system for active participants
- Federation-aware topic indexing and search

## Getting Started

### First-Time Setup
```bash
# Complete development environment setup
task dev-setup
```

### Development Workflow
```bash
# Start new feature from GitHub issue
task worktree-dev ISSUE=123

# Daily development cycle
task dev-check                 # Generate code + lint + test
task commit-check              # Validate before committing
task pr-create                 # Submit pull request

# After PR merge
task worktree-cleanup BRANCH=issue-123
```

### Running the Application
```bash
# Start the server (multiple options)
task run                       # Using Taskfile (recommended)
go run main.go                 # Direct Go execution
make run                       # Using Makefile
```

### Getting Help
```bash
task welcome                   # Welcome message and quick start
task quickstart                # Interactive quickstart guide
task help                      # Comprehensive help (opens in pager)
task --list                    # Show all available tasks
task help-claude               # Commands for Claude Code assistance
```

## Working with Claude Code

This project is optimized for collaboration with [Claude Code](https://claude.ai/code). For best results:

### Starting a Session
```bash
task claude-context           # Provide complete project context to Claude Code
```

### Getting Help
```bash
task project-health           # Comprehensive project status
task docs-status              # Check documentation freshness
task help-claude              # Show Claude Code assistance commands
```

### Development Best Practices
- Use `task` commands instead of manual git/tool operations
- Run `task dev-check` before committing
- Keep `CLAUDE.md` as the authoritative development guide
- Use GitHub issues and worktrees for feature development

See `CLAUDE.md` for complete development workflow documentation.
