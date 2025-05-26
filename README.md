# dis.quest — Go POC for ATProtocol Discussions

This project is an experimental implementation of a secure, discussion-style messaging protocol using [ATProtocol](https://atproto.com/) and optionally [OpenTDF](https://github.com/virtru/OpenTDF).

## Features

- Custom lexicons under `quest.dis.*` and `quest.dis.sec.*`
- Discussion topics, messages, participation tracking
- Optional encrypted messaging using OpenTDF-style fields
- Targeting full PDS/AppView compatibility
- Written in idiomatic Go

## Getting Started

```bash
go run main.go
