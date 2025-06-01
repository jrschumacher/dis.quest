# Makefile for dis.quest

.PHONY: run test lint format build clean

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run

format:
	gofmt -w . && goimports -w .

build:
	go build -o bin/disquest ./cmd/server

clean:
	rm -rf bin/