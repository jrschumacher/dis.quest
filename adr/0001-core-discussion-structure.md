---
id: 0001
title: Core Discussion Structure for Dis.quest
status: accepted
date: 2025-06-06
tags:
  - adr
  - protocol
  - discussions
  - atproto
lexicons:
  - quest.dis.topic
  - quest.dis.post
  - quest.dis.participation
supersedes: null
supersededBy: null
---

# ADR-0001: Core Discussion Structure for Dis.quest

## Context
ATProtocol natively supports posts and chats but lacks a first-class model for **structured, topic-based discussions**. For many use cases—community forums, long-lived Q&A, or asynchronous collaboration—this limitation introduces friction and fragmentation. Dis.quest aims to fill this gap by introducing a discussion-centric data model.

The goal is to support persistent, organized conversations that can be rendered in clients and app views in a way that feels familiar to users of platforms like GitHub Discussions or Discourse, while remaining decentralized and compatible with ATProtocol principles.

## Decision
We will define a **core discussion model** using custom ATProto lexicons under the `quest.dis.*` namespace.

### Key Record Types

- `quest.dis.topic`
  Represents a discussion topic. Includes:
  - `title` (string): Human-readable summary
  - `body` (string): Initial message
  - `category` (optional string): Classification label
  - `selectedAnswer` (optional CID): Q&A mode supported
  - `createdBy`, `createdAt`, `updatedAt`

- `quest.dis.post`
  A message or reply posted within a topic. Includes:
  - `topic` (CID or URI): Reference to the parent topic
  - `body` (string): The post content
  - `replyTo` (optional CID): Enables shallow threading
  - `reactions` (map): Emoji reactions (e.g. 👍 → count)
  - `upvotes` (int): Endorsement signal

- `quest.dis.participation`
  Tracks a user's opt-in to follow, moderate, or otherwise engage with a topic. Enables clients to render “who’s watching” or notify followers.

### Constraints

- No deep recursive threading (flat topic → post → reply)
- One topic per thread root
- All posts must reference a valid topic
- Participation records are optional and do not imply permissions

### App View Support
The model will be compatible with PDS and AppView behavior, allowing:
- Feed-like topic lists
- Post aggregation by topic
- Optional display of followers, reactions, or Q&A state

## Consequences

### Positive
- Enables structured, durable conversations on ATProtocol
- Aligns with common mental models (forums, threads, Q&A)
- Composable lexicons can be reused by other apps
- Minimal set of records allows progressive enhancement

### Negative
- Adds a layer of abstraction over native post/feed behavior
- Requires custom clients or app views for full UX
- Flat reply structure may not be sufficient for deeply-nested discussions

## Alternatives Considered
- **Reusing `app.bsky.feed.post`**: Lacks structured relationship modeling and would require non-standard conventions.
- **Using chat semantics**: Inappropriate for non-ephemeral, scoped conversations.
- **Federated groups model first**: Adds unnecessary complexity for MVP.

## Related Work
- [Discourse](https://www.discourse.org/)
- [GitHub Discussions](https://docs.github.com/en/discussions)
- ATProto `app.bsky.feed.post` and `app.bsky.graph.list`

## Next Steps
- Implement `quest.dis.topic`, `quest.dis.post`, `quest.dis.participation` lexicons
- Create reference Go implementation for publishing and reading discussions
- Begin app view development for topic browsing and reply aggregation
