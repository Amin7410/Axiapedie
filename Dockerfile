# Stage 1: Build binary
FROM golang:1.23-alpine AS builder

# Install build tools for CGO (required by mattn/go-sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled
ENV CGO_ENABLED=1
RUN go build -ldflags="-w -s" -o server ./cmd/server

# Stage 2: Runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy binary and static assets
COPY --from=builder /app/server .
COPY --from=builder /app/db ./db
COPY --from=builder /app/web ./web

# Create empty uploads and data directories if they don't exist
RUN mkdir -p /app/uploads /app/data

EXPOSE 8081

CMD ["./server"]
