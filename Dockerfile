# Multi-stage Dockerfile untuk Flash Sale API
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Salin source code
COPY . .

# Build binary utama (API Server) dan Worker Daemon
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/worker-bin ./cmd/worker/main.go

# Minimal runtime image
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server /app/server
COPY --from=builder /app/worker-bin /app/worker-bin
COPY --from=builder /app/.env.example /app/.env.example

EXPOSE 3000

CMD ["/app/server"]
