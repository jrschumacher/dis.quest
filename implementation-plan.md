# dis.quest Implementation Plan

This plan outlines the architecture and development roadmap for `dis.quest`, a secure discussion platform built on the ATProtocol with optional OpenTDF-based encryption. Each item is tracked with checkboxes to provide persistent progress tracking.

## ⚙️ Phase 0: Dev Environment & Bootstrap

- [ ] Setup Go module and folder structure
- [ ] Add Makefile or dev scripts for common commands
- [ ] Add `.env.example` and configuration loader
- [ ] Add CI (e.g., GitHub Actions with `golangci-lint`, tests)
- [ ] Add local testing support for PDS stubs or mocks

## ✅ Phase 1: UI/UX Foundation

- [ ] Draft landing page (overview of the platform)
- [ ] Draft login page (OAuth2-based auth)
- [ ] Draft discussion page (topic threads, messages, reply UI)

## 🔐 Phase 2: Authentication

- [ ] Implement OAuth2 login using Bluesky/ATProto identity
- [ ] Store user sessions securely (cookies or token-based)
- [ ] Enable reading/writing to user's PDS after authentication

## 🧠 Phase 3: Server + Lexicon Support

- [ ] Implement backend API in Go to:
  - [ ] Serve lexicons (custom `quest.dis.*` definitions)
  - [ ] Create a new post via ATProto
  - [ ] Handle post signing and encoding
- [ ] Add TDF-based policy support to backend post creation

## 🔁 Phase 4: Firehose Integration

- [ ] Connect to ATProto firehose
- [ ] Parse relevant `quest.dis.*` events
- [ ] Store incoming posts
- [ ] Deduplicate and validate message integrity

## 🗄️ Phase 5: Storage and Indexing

- [ ] Choose persistence backend (PostgreSQL, SQLite, or BadgerDB)
- [ ] Model discussion threads and messages
- [ ] Store user profiles for mapping DID → handle
- [ ] Add indexing for efficient post linking and retrieval

## 📡 Phase 6: Realtime Updates

- [ ] Add websocket server support
- [ ] Push updates on new post events
- [ ] Push notifications for mentions or new messages in threads

## 🧩 Phase 7: REST Interfaces

- [ ] Build internal REST APIs to fetch messages by thread ID
- [ ] Support posting messages locally and syncing to PDS asynchronously
- [ ] Expose metadata endpoints (e.g., latest topics, participants)

---

This doc is maintained as the source of truth for planning and tracking. Let me know when you’d like to assign priorities, owners, or split tasks into subtickets.
# dis.quest Implementation Plan

This plan outlines the architecture and development roadmap for `dis.quest`, a secure discussion platform built on the ATProtocol with optional OpenTDF-based encryption. Each item is tracked with checkboxes to provide persistent progress tracking.

## ⚙️ Phase 0: Dev Environment & Bootstrap

- [ ] Setup Go module and folder structure
- [ ] Add Makefile or dev scripts for common commands
- [ ] Add `.env.example` and configuration loader
- [ ] Add CI (e.g., GitHub Actions with `golangci-lint`, tests)
- [ ] Add local testing support for PDS stubs or mocks

## ✅ Phase 1: UI/UX Foundation

- [ ] Draft landing page (overview of the platform)
- [ ] Draft login page (OAuth2-based auth)
- [ ] Draft discussion page (topic threads, messages, reply UI)

## 🔐 Phase 2: Authentication

- [ ] Implement OAuth2 login using Bluesky/ATProto identity
- [ ] Store user sessions securely (cookies or token-based)
- [ ] Enable reading/writing to user's PDS after authentication

## 🧠 Phase 3A: Server API + Lexicon Support

- [ ] Implement REST API in Go to:
  - [ ] Serve lexicons (custom `quest.dis.*` definitions)
  - [ ] Create stub endpoint for post creation (local echo or log)
- [ ] Add unit tests for lexicon structure and API responses

## ✍️ Phase 3B: Post Creation via ATProto

- [ ] Implement write to ATProto (authenticated user)
- [ ] Handle signing, compression, and schema mapping
- [ ] Integrate OpenTDF policy wrapping

## 🧩 Phase 4: OpenTDF Policy Tooling

- [ ] Define policy template support (`io.opentdf.tdf.template`)
- [ ] CLI or UI for creating and managing attributes
- [ ] Debug tool for inspecting CEKs, KAS metadata

## 🗄️ Phase 5: Storage and Indexing

- [ ] Choose persistence backend (PostgreSQL, SQLite, BadgerDB)
- [ ] Model discussion threads, messages, and linkage
- [ ] Store user profiles for DID ↔ handle mapping
- [ ] Add indexing for thread lookup and retrieval

## 🔁 Phase 6: Firehose Integration

- [ ] Connect to ATProto firehose
- [ ] Parse `quest.dis.*` events
- [ ] Store incoming posts into the backend
- [ ] Deduplicate and validate signature/integrity

## 📡 Phase 7: Realtime Notifications

- [ ] Add websocket server support
- [ ] Push updates for new messages in active threads
- [ ] Implement mention detection and topic-level alerts

## 🌐 Phase 8: REST Interfaces

- [ ] Build internal REST APIs to:
  - [ ] Fetch messages by thread ID
  - [ ] Post new messages locally and sync to PDS
  - [ ] Query metadata (latest topics, participants)

## 🔍 Phase 9: Observability & Monitoring

- [ ] Add structured logging and tracing (e.g., zap, slog)
- [ ] Implement health checks
- [ ] Add metrics (Prometheus/Grafana compatibility)
- [ ] Add retry queue for post sync failures

---

This doc is maintained as the source of truth for planning and tracking. Let me know when you’d like to prioritize, assign owners, or split tasks further.