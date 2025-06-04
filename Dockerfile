# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

RUN apk update && apk add --no-cache git curl

WORKDIR /app

# Install tools for building
RUN go install github.com/go-task/task/v3/cmd/task@latest && \
    go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN task build


# Stage 2: Create the smallest possible image
FROM scratch
WORKDIR /app

# Use a static non-root user if available, otherwise fallback to root
# USER nonroot:nonroot

COPY --from=builder /app/bin/disquest .

EXPOSE 3000

CMD ["./disquest", "start"]

