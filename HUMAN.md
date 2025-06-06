# Claude Code Prompting Cheatsheet

## 🚀 Core Prompting Principles

### Be Explicit About Scope
```
✅ "Refactor the authentication logic in internal/auth/ without touching the database models"
❌ "Clean up the auth code"
```

### Set Clear Context
```
✅ "This is a Go project with Templ templates, SQLC, and ATProtocol integration"
❌ "Fix this Go app"
```

### Provide Project Context First
```
✅ "Run `task claude-context` first, then implement OAuth refresh token handling"
❌ "Add refresh token support"
```

### Break Down Complex Tasks
```
✅ "1. Analyze current API endpoints, 2. Identify missing error handling, 3. Add comprehensive error handling with logging"
❌ "Make the API better"
```

## 🏷️ Hashtag Shortcuts (Creative Modifiers)

Add these to the end of any prompt for specific behaviors:

### Analysis & Understanding
- `#explain` - Provide detailed explanations of changes made
- `#analyze` - Deep dive into code structure and patterns before acting
- `#document` - Generate comprehensive documentation for changes
- `#breakdown` - Show step-by-step reasoning for complex operations

### Verification & Testing
- `#verify` - Run tests and validate changes after implementation
- `#test` - Create or update tests alongside code changes
- `#lint` - Check code style and fix linting issues
- `#benchmark` - Measure performance impact of changes

### Safety & Caution
- `#safe` - Make minimal, conservative changes with backups
- `#preview` - Show what will be changed before executing
- `#isolate` - Work only in specified directories/files
- `#backup` - Create backups before making changes

### Style & Quality
- `#clean` - Focus on code cleanliness and readability
- `#optimize` - Prioritize performance improvements
- `#consistent` - Enforce existing project patterns and conventions
- `#modern` - Use latest best practices and language features

### Output Control
- `#quiet` - Minimal output, just show results
- `#verbose` - Detailed logging of all actions taken
- `#summary` - Provide executive summary of changes made
- `#diff` - Show clear before/after comparisons

## 🎯 dis.quest Specific Use Cases

### ATProtocol Integration
```
"Add a new lexicon for topic voting under quest.dis.vote namespace #explain #verify #document"
```

### Templ Template Work
```
"Create a new component for displaying discussion threads in components/ #consistent #preview"
```

### Authentication Enhancement
```
"Improve OAuth token refresh handling in internal/auth/ #safe #test #explain"
```

### Database Migration
```
"Add new migration for topic categories table #backup #verify #document"
```

### API Handler Development
```
"Add new handlers for topic management in server/app/ following existing patterns #consistent #test"
```

### Code Review Preparation
```
"Run task dev-check and prepare this branch for PR #clean #lint #summary"
```

## 🎯 General Use Cases

### Feature Implementation
```
"Add dark mode toggle to the UI components #consistent #document #preview"
```

### Bug Fixing
```
"Fix the session cookie issue in auth handlers #analyze #test #verify"
```

### Legacy Code Modernization
```
"Update deprecated ioutil usage to os package #modern #safe #diff"
```

## 🔧 Advanced Patterns

### Conditional Logic
```
"If the test coverage is below 80%, write additional tests; otherwise, just refactor for readability #verify"
```

### Multi-Stage Operations
```
"Phase 1: Analyze current architecture #analyze
Phase 2: Propose improvements #preview
Phase 3: Implement changes #safe #verify"
```

### Context-Aware Commands
```
"Following our established patterns in /components/ui/, create a new modal component #consistent #document"
```

## 📋 Quick Reference Templates

### New Feature
```
"Implement [FEATURE] in [LOCATION] following [PATTERN] #explain #test #consistent"
```

### Bug Investigation
```
"Debug [ISSUE] in [FILE/FUNCTION] #analyze #verbose #verify"
```

### Code Cleanup
```
"Clean up [TARGET] focusing on [ASPECTS] #clean #lint #summary"
```

### Performance Optimization
```
"Optimize [COMPONENT] for [METRIC] #benchmark #explain #verify"
```

### Documentation
```
"Document [CODE/API] with examples and usage patterns #document #explain"
```

## 💡 dis.quest Pro Tips

1. **Start with context**: Run `task claude-context` before complex requests
2. **Use project workflows**: Reference `task dev-check`, `task worktree-dev ISSUE=123`, etc.
3. **Follow project patterns**: Mention "following existing auth handlers" or "using SQLC patterns"
4. **Be specific about scope**: "internal/auth/", "server/app/", "components/" are key directories
5. **Leverage task automation**: "Run task dev-check after changes" instead of manual steps
6. **Reference lexicons**: Mention `quest.dis.*` namespace for ATProtocol work
7. **Consider documentation**: CLAUDE.md updates often needed after workflow changes

## 💡 General Pro Tips

1. **Combine hashtags thoughtfully**: `#safe #preview #explain` for critical changes
2. **Use project-specific context**: Reference your actual file structures and naming conventions
3. **Be specific about constraints**: Mention deployment environments, browser support, etc.
4. **Request explanations for learning**: `#explain` helps you understand the changes
5. **Always verify important changes**: `#verify #test` for production-bound code

## 🚨 Safety Hashtags for Critical Code

- `#production` - Extra caution for production environments
- `#database` - Special care with database operations
- `#security` - Focus on security implications
- `#breaking` - Acknowledge potential breaking changes
- `#rollback` - Prepare rollback procedures

## 📋 dis.quest Quick Start Template

```
"task claude-context

[Your specific request here]

Following dis.quest patterns in [directory], [action] #[relevant hashtags]"
```

### Example Session Start
```
"task claude-context

I need to add user profile management to the discussion platform. 

Following existing auth patterns in internal/auth/ and server/auth-handlers/, create profile CRUD operations with proper OAuth validation #explain #consistent #test"
```

---

*Remember: Claude Code operates in your real development environment. Always review changes, especially for critical systems! Use `task dev-check` before committing.*
