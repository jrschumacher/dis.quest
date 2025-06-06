---
id: secure
title: Dis.quest Secure Extension – Encrypted Discussions with OpenTDF
description: Product brief for integrating OpenTDF-style encryption into the Dis.quest protocol.
status: Draft
lastUpdated: 2025-06-06
version: 0.9
lexicons:
  - quest.dis.sec.post
  - quest.dis.sec.policy
  - quest.dis.sec.kas
  - quest.dis.sec.embed
tags:
  - product
  - encryption
  - opentdf
  - secure-messaging
layout: doc
---

# 🔐 Product Brief: Dis.quest Secure Extension – Encrypted Discussions with OpenTDF

## Overview
The **Dis.quest Secure Extension** introduces optional support for **encrypted discussions** within the Dis.quest protocol using [OpenTDF](https://github.com/virtru/OpenTDF). This layer adds **attribute-based access control (ABAC)** to topics, posts, and media attachments, enabling **confidential and policy-enforced communication** on top of ATProtocol.

Messages and attachments remain private unless the recipient meets the defined access policy. This allows for confidential group threads, secure data sharing, and fine-grained access control—all within a decentralized discussion framework.

## Goals
- Enable encrypted messages and media in ATProtocol discussions
- Ensure confidentiality with no content leakage (including blob CIDs)
- Support **shared policy** enforcement across a topic/thread
- Maintain compatibility with the core Dis.quest app view and lexicons
- Demonstrate a reference model for extending ATProtocol with protected content

## Design Principles

- **Unified Policy Model**: All posts in a topic share a single OpenTDF policy
- **Shared CEK**: Topic and embeds use a common encryption key
- **No CID Leakage**: Blob references are wrapped and encrypted
- **Extensible Fields**: Secure fields are stored alongside their plaintext equivalents using a dedicated `quest.dis.sec.*` namespace

## Lexicon Overview

| Lexicon ID             | Purpose                                    |
| ---------------------- | ------------------------------------------ |
| `quest.dis.sec.post`   | Encrypted post body and optional signature |
| `quest.dis.sec.policy` | Compressed ABAC policy definition          |
| `quest.dis.sec.kas`    | KAS metadata (URL, protocol, bindings)     |
| `quest.dis.sec.embed`  | Encrypted blob reference wrapper           |

## Secure Post Record Structure

| Field   | Description                                             |
| ------- | ------------------------------------------------------- |
| `pl`    | Base64-encoded, encrypted post payload                  |
| `po`    | Base64-encoded OpenTDF policy blob                      |
| `ka`    | Compressed list of KAS coordination metadata            |
| `sg`    | Optional detached signature                             |
| `embed` | Reference to encrypted blob (via `quest.dis.sec.embed`) |

## Use Cases
- Secure group discussions with dynamic access control
- Protected media/file sharing inside topics
- Confidential forums within federated networks
- Policy-scoped knowledge bases (e.g. internal Q&A)

## Scope

### ✅ In Scope
- Optional encryption of topic messages
- Shared encryption policy across topic/thread
- Encrypted blob embedding via ATProto blob store
- Attribute-based access policies with KAS coordination
- Go implementation of TDF encoding/decoding for secure fields
- Compatibility with Dis.quest unprotected app view

### 🚫 Out of Scope
- Per-message policy overrides
- Mixed protection modes within a thread
- Encrypted identity or handle masking
- End-to-end encryption across PDS boundaries (for now)

## Integration Strategy

- Secure fields are **optional overlays** to base lexicons (`quest.dis.post`)
- Clients supporting OpenTDF can decrypt and render protected content
- Non-supporting clients degrade gracefully (e.g. "Encrypted post" placeholder)
- Policy evaluation is delegated to client-side or trusted KAS

## Future Directions
- Per-participant policy targeting and revocation
- Transparent KAS discovery and federation
- Support for encrypted mentions and metadata
- Encrypted reactions and moderation signals

