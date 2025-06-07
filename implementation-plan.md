# dis.quest Implementation Plan

This plan outlines the architecture and development roadmap for `dis.quest`, a structured discussion platform built on ATProtocol with topic-post-participation modeling. Each item is tracked with checkboxes to provide persistent progress tracking.

## Current Status Overview

**✅ COMPLETED**: Development tooling, database foundation, authentication infrastructure, basic UI templates, database integration & core data layer  
**🚧 IN PROGRESS**: Core discussion functionality, ATProtocol integration  
**⏭️ NEXT**: Complete authentication flow, implement core discussion workflows

---

## ⚙️ Phase 0: Development Environment & Infrastructure

- [x] Setup Go module and project structure with domain-organized handlers
- [x] Add comprehensive Taskfile automation for all development workflows
- [x] Add configuration management with Viper (environment-aware, validation, secrets)
- [x] Add CI/CD with GitHub Actions (lint, test, conventional commits)
- [x] Add git hooks with Lefthook for automated quality checks
- [x] Add database architecture with SQLC type-safe queries
- [x] Add template system with Templ for server-side rendering
- [ ] Add local testing support for PDS stubs or mocks
- [ ] Add integration tests for authentication flow

## 🎨 Phase 1: UI/UX Foundation

- [x] Landing page with project overview and OAuth login
- [x] Login page with OAuth2 redirect to Bluesky
- [x] Discussion page mockup with static threading UI
- [x] Basic Pico CSS styling and responsive layout
- [ ] Connect templates to real data from database
- [ ] Add form handling for topic/message creation
- [ ] Add error handling and user feedback UI
- [ ] Add user authentication state in templates

## 🔐 Phase 2: Authentication & User Management

- [x] OAuth2 infrastructure with PKCE and DPoP support
- [x] Secure session cookie management (environment-aware)
- [x] State token generation for CSRF protection
- [x] PDS discovery endpoint (currently hardcoded to Bluesky)
- [ ] **CRITICAL**: Complete OAuth token validation and JWT parsing
- [ ] **CRITICAL**: Implement user session to DID mapping
- [ ] Add session persistence and user context middleware
- [ ] Add proper logout flow with session cleanup
- [ ] Add token refresh mechanism
- [ ] Enable reading/writing to user's PDS after authentication

## 🗄️ Phase 3: Database Integration & Core Data Layer

- [x] Database schema for discussions (topics, messages, participation)
- [x] SQLC queries for all CRUD operations with proper indexing
- [x] Migration system with Tern
- [x] Type-safe models with JSON serialization
- [x] **CRITICAL**: Connect database queries to HTTP handlers
- [x] **CRITICAL**: Implement user session to database mapping
- [x] Add database transaction support for complex operations
- [x] Add proper error handling and validation
- [x] Add data access layer abstraction (SQLC provides sufficient abstraction)

## 🧠 Phase 4: Core Discussion Functionality

- [ ] **HIGH PRIORITY**: Topic creation and management
  - [ ] POST /topics endpoint with form handling
  - [ ] Topic listing with pagination and filtering
  - [ ] Topic detail view with metadata
  - [ ] Category support and filtering
- [ ] **HIGH PRIORITY**: Message posting and threading
  - [ ] POST /topics/{id}/messages endpoint
  - [ ] Reply functionality with shallow threading (replyTo)
  - [ ] Message editing and deletion
  - [ ] Message validation and sanitization
- [ ] **MEDIUM PRIORITY**: Participation tracking
  - [ ] Follow/unfollow topics
  - [ ] Participation signals (active/inactive)
  - [ ] User notification preferences
- [ ] **MEDIUM PRIORITY**: Q&A functionality
  - [ ] selectedAnswer support for topics
  - [ ] Answer selection by topic creators
  - [ ] Q&A mode UI indicators

## 🌐 Phase 5: ATProtocol Lexicon Integration

- [x] Lexicon definitions for quest.dis.topic, quest.dis.post, quest.dis.participation
- [ ] **HIGH PRIORITY**: ATProtocol record creation using custom lexicons
- [ ] **HIGH PRIORITY**: ATProtocol record reading from user PDS
- [ ] Lexicon validation and schema enforcement
- [ ] PDS write operations for topic/message creation
- [ ] PDS read operations for data retrieval
- [ ] Handle ATProtocol CID references and blob storage
- [ ] Implement proper ATProto signing and compression

## 🎯 Phase 6: Enhanced Features

- [ ] **MEDIUM PRIORITY**: Reactions and upvotes
  - [ ] Emoji reaction system with counts
  - [ ] Upvote functionality for posts
  - [ ] Reaction aggregation and display
- [ ] **MEDIUM PRIORITY**: Search and discovery
  - [ ] Full-text search across topics and messages
  - [ ] Category-based browsing
  - [ ] User activity feeds
- [ ] **LOW PRIORITY**: Attachments and media
  - [ ] File upload to ATProto blob store
  - [ ] Image embedding and preview
  - [ ] Attachment security validation

## 🔄 Phase 7: Real-time Features & Synchronization

- [ ] Firehose integration for ATProtocol events
  - [ ] Connect to ATProto firehose
  - [ ] Parse quest.dis.* records from stream
  - [ ] Store incoming posts and maintain consistency
  - [ ] Handle deduplication and signature validation
- [ ] WebSocket support for real-time updates
  - [ ] Live updates for active discussions
  - [ ] Real-time participation signals
  - [ ] Push notifications for followed topics

## 🔒 Phase 8: Security Extensions (Optional)

**Note**: OpenTDF encryption is an optional extension documented in `/docs/prd/secure.md`

- [ ] OpenTDF policy framework
  - [ ] Policy template support
  - [ ] Attribute-based access control (ABAC)
  - [ ] KAS (Key Access Service) integration
- [ ] Secure lexicon implementation (quest.dis.sec.*)
  - [ ] Encrypted post content
  - [ ] Policy attachment to topics
  - [ ] Secure blob references
- [ ] Client-side encryption/decryption
  - [ ] TDF encoding/decoding in Go
  - [ ] Policy evaluation
  - [ ] Graceful degradation for non-supporting clients

## 📊 Phase 9: Production Readiness

- [x] Structured logging infrastructure
- [x] Health check endpoints
- [ ] Metrics and monitoring (Prometheus compatibility)
- [ ] Rate limiting and abuse prevention
- [ ] Performance optimization and caching
- [ ] Deployment automation (Docker, CI/CD)
- [ ] Error tracking and alerting
- [ ] Backup and disaster recovery

## 🧪 Phase 10: Testing & Quality Assurance

- [x] Unit test infrastructure
- [x] Linting and code quality automation
- [ ] Integration tests for authentication flow
- [ ] End-to-end tests for discussion workflows
- [ ] Performance testing and benchmarking
- [ ] Security testing and vulnerability assessment
- [ ] Load testing for concurrent users

---

## 🎯 Immediate Next Steps (Current Sprint)

### Week 1: Core Functionality Foundation
1. **Complete authentication flow**: JWT validation, user session management
2. **Connect database to handlers**: Implement real topic/message CRUD operations
3. **Basic discussion workflows**: Topic creation, message posting, threading

### Week 2: ATProtocol Integration
1. **Implement lexicon operations**: Create/read ATProtocol records
2. **PDS integration**: Write topic/message data to user PDS
3. **User experience improvements**: Error handling, form validation, UI polish

### Week 3: Enhanced Features
1. **Participation system**: Follow/unfollow topics, notification preferences
2. **Q&A functionality**: selectedAnswer support and UI
3. **Search and discovery**: Basic search, category filtering

---

## Architecture Alignment Notes

**✅ Implemented and Aligned**:
- Database-first approach with type-safe queries matches ADR-0001 constraints
- Flat threading model (topic → post → reply) implemented in schema
- Custom lexicon namespace (quest.dis.*) properly defined
- Development workflow matches CLAUDE.md specifications exactly

**🚧 Partially Aligned**:
- Lexicon structure matches PRD specifications but not yet implemented in code
- Handler organization follows documented patterns but lacks business logic
- Authentication infrastructure exists but needs completion

**⚠️ Gaps to Address**:
- No actual ATProtocol integration beyond authentication scaffolding
- Missing user context propagation from authentication to data operations
- Templates exist but not connected to real data

This plan prioritizes completing the core discussion functionality first, then progressively enhancing with ATProtocol integration and advanced features. The optional OpenTDF encryption remains as a future extension without blocking core development.