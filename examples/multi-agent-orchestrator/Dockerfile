# Multi-Agent Orchestrator Dockerfile
# Multi-stage build for minimal image size

# =============================================================================
# BUILD STAGE
# =============================================================================
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /orchestrator \
    ./examples/multi-agent-orchestrator

# =============================================================================
# RUNTIME STAGE
# =============================================================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' orchestrator
USER orchestrator

WORKDIR /app

# Copy binary from builder
COPY --from=builder /orchestrator .

# Copy configuration
COPY --chown=orchestrator:orchestrator examples/multi-agent-orchestrator/agents.yaml ./config/

# Create directories for persistent storage
RUN mkdir -p /app/storage

# Environment variables
ENV CONFIG_PATH=/app/config/agents.yaml
ENV OLLAMA_HOST=http://ollama:11434
ENV MEMORY_STORAGE_DIR=/app/storage

# Health check (requires Ollama to be available)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Expose port (if HTTP API is added later)
EXPOSE 8080

# Run the orchestrator
ENTRYPOINT ["./orchestrator"]
