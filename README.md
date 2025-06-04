# dis.quest — Go POC for ATProtocol Discussions

This project is an experimental implementation of a secure, discussion-style messaging protocol using [ATProtocol](https://atproto.com/) and optionally [OpenTDF](https://github.com/virtru/OpenTDF).

## Features

- Custom lexicons under `quest.dis.*` and `quest.dis.sec.*`
- Discussion topics, messages, participation tracking
- Optional encrypted messaging using OpenTDF-style fields
- Q&A support via `selectedAnswer` on topics
- Targeting full PDS/AppView compatibility
- Written in idiomatic Go

## Project Background and Purpose

`dis.quest` is a prototype implementation of a discussion-based messaging platform built on ATProtocol. It is designed to model GitHub-style discussions (topics, replies, participation), while optionally supporting encrypted message content using OpenTDF. The project is exploring how decentralized, attribute-based access control (ABAC) can integrate with federated discussion threads.

Key architectural ideas include:
- Each post is a custom record (e.g., `quest.dis.post`) with optional OpenTDF-protected payloads
- All messages in a thread can be encrypted under a unified OpenTDF policy
- Embeds (e.g., media blobs) are encrypted and referenced using the ATProto blob store
- Custom lexicons define topics, messages, participation, and security contexts

While OpenTDF integration is being explored as part of this prototype, it is not a core feature of the base discussion experience. Instead, it serves as a demonstration of how the ATProtocol and its lexicon system can be extended to support encrypted payloads and attribute-based access control. This includes the ability to define and use additional lexicons that go beyond the app-specific view, opening the door to broader interoperability and experimentation.

## Scope

This project is intentionally limited in scope to prioritize core discussion features and integration with ATProtocol and OpenTDF.

### ✅ In Scope

- **Topic** — Represents a discussion subject with an initial message and metadata
- **Topic Category** — Grouping or classification (e.g., hashtags) to organize topics
- **Participation** — Opt-in to follow a topic and receive updates or reply notifications
- **Top-Level Message** — A direct response to a topic
- **Thread** — Replies to a top-level message, supporting conversational depth
- **Emoji Reactions** — Non-verbal message responses (e.g., 👍 ❤️ 🔥)
- **Upvotes** — Lightweight endorsement signal for a message or thread
- **Q&A Topics** — Threads with an accepted answer (stored in `selectedAnswer`)
- **Mentions (`@handle`)** — Notifies tagged users (planned)
- **WebSocket subscriptions** — Realtime updates for followed discussions (planned)
- **Attachments** — Encrypted media or files associated with a post
- **Defined Categories** — Predefined list of categories for browsing and filtering topics

### 🚫 Out of Scope

- **Polls** — Voting-based discussions
- **Direct Messaging** — 1:1 or group chat support
- **Private Threads** — Topics with per-message visibility restrictions
- **Rich Text Editors** — Markdown/WYSIWYG formatting is deferred to the client layer

## Getting Started

```bash
# Start the local dev server
go run main.go

# (Optional) Generate lexicons
# Placeholder command — replace with actual generator when available
echo "Generating lexicons..."
```