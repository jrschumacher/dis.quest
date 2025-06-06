---
id: core
title: Dis.quest – Structured Discussions on ATProtocol
description: Product brief for the base discussion protocol using ATProtocol with topic-post-participation modeling.
status: Final
lastUpdated: 2025-06-06
version: 1.0
lexicons:
  - quest.dis.topic
  - quest.dis.post
  - quest.dis.participation
tags:
  - product
  - atproto
  - discussions
  - decentralized
layout: doc
---

# 🧭 Product Brief: Dis.quest – Structured Discussions on ATProtocol

## Overview
**Dis.quest** is a structured, topic-based discussion protocol and app view built on top of [ATProtocol](https://atproto.com/). It introduces a GitHub Discussions–style model for decentralized communities, enabling asynchronous conversation around shared topics—scoped by participation, threaded replies, reactions, and optional Q&A flow.

The base protocol is designed to function independently with clear separation between public messaging and optional protected content extensions.

## Goals
Dis.quest aims to define a protocol-native way to:
- Start and organize **topics** scoped around a defined subject
- Support **replies** and **threads** in a structured conversation
- Track **participation** and notify followers of updates
- Enable **optional Q&A**, **reactions**, and **app-level categorization**
- Provide an idiomatic, reference implementation in **Go**

## Design Principles

### Core Discussion Model
- **Topics**: Root discussion containers with title, body, and categorization
- **Posts**: Messages and replies that reference a topic with shallow threading support
- **Participation**: User follow and moderation signals for topics
- **Flat Threading**: Constrained to topic → post → reply to avoid deep nesting complexity

### ATProtocol Native
- Custom lexicons under `quest.dis.*` namespace
- Compatible with PDS and AppView patterns
- Leverages ATProto blob store for attachments
- Maintains decentralized identity and data ownership

## Lexicon Overview

| Lexicon ID                | Purpose                          |
| ------------------------- | -------------------------------- |
| `quest.dis.topic`         | Defines a discussion topic       |
| `quest.dis.post`          | Contribution post or reply       |
| `quest.dis.participation` | Follow or moderation signal      |

## Core Record Structures

### quest.dis.topic
| Field           | Description                                    |
| --------------- | ---------------------------------------------- |
| `title`         | Human-readable topic summary                   |
| `body`          | Initial topic description/content              |
| `category`      | Optional classification label                  |
| `selectedAnswer`| Optional CID reference for Q&A mode           |
| `createdBy`     | Author DID                                     |
| `createdAt`     | ISO timestamp                                  |
| `updatedAt`     | Last modification timestamp                    |

### quest.dis.post
| Field      | Description                                       |
| ---------- | ------------------------------------------------- |
| `topic`    | CID or URI reference to parent topic             |
| `body`     | Post content                                      |
| `replyTo`  | Optional CID for shallow threading               |
| `reactions`| Map of emoji reactions to counts                 |
| `upvotes`  | Endorsement signal count                         |
| `embed`    | Optional blob reference for attachments          |

### quest.dis.participation
| Field        | Description                                     |
| ------------ | ----------------------------------------------- |
| `topic`      | CID or URI reference to topic                  |
| `type`       | Participation type (follow, moderate, etc.)    |
| `active`     | Boolean signal for current participation       |

## Key Features

- Topics with title, description, and categories
- Posts scoped to a topic, supporting threaded replies
- Participation model for follows and moderation
- Q&A topics with `selectedAnswer` support
- Emoji reactions and upvotes
- Mentions (`@handle`) for user notifications *(planned)*
- Attachments using ATProto blob store
- Defined categories for filtering/browsing

## Use Cases
- Community forums and discussion boards
- Q&A knowledge bases with selected answers
- Project-specific threaded discussions
- Asynchronous team collaboration
- Public discourse with structured threading

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
- Deep recursive threading

## Integration Strategy

### PDS Compatibility
- Each user's discussion data stored in their own PDS
- Public topic metadata shared across app views
- User-scoped participation and interaction records

### App View Support
- Feed-like topic lists and aggregation
- Post threading and reply organization
- Reaction and participation summaries
- Search and categorization indexing

### Client Implementation
- Server-side rendering with minimal JavaScript
- Progressive enhancement for interactivity
- Graceful degradation for unsupported features

## Future Directions
- Moderation roles and admin actions
- Nested threading and reply hierarchies
- Reaction badges and reputation signals
- Notification system for active participants
- Federation-aware topic indexing and search
- Optional encryption extensions (see secure.md)

## Related Work
- [GitHub Discussions](https://docs.github.com/en/discussions)
- [Discourse](https://www.discourse.org/)
- ATProto `app.bsky.feed.post` and `app.bsky.graph.list`
- [ADR-0001: Core Discussion Structure](/adr/0001-core-discussion-structure.md)