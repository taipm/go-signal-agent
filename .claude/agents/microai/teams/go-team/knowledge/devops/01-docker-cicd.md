# Docker & CI/CD - DevOps Agent Knowledge

**Version:** 1.0.0
**Agent:** DevOps Agent

---

## TL;DR

- Multi-stage Docker builds cho small images
- GitHub Actions cho CI/CD
- Version embedding với ldflags
- Distroless/scratch base images
- Makefile cho common tasks

---

## 1. Multi-Stage Dockerfile

### Standard Go Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION:-dev}" \
    -o /app/server ./cmd/api

# Final stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/server /server

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### With CA Certificates

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/server ./cmd/api

# Final stage
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server

USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### Development Dockerfile

```dockerfile
# Development with hot reload
FROM golang:1.21

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]
```

---

## 2. GitHub Actions CI/CD

### Complete Workflow

```yaml
# .github/workflows/ci.yml
name: CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  release:
    types: [created]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run tests
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Build
        run: |
          CGO_ENABLED=0 go build \
            -ldflags="-w -s -X main.version=${{ github.sha }}" \
            -o ./bin/server ./cmd/api

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: server
          path: ./bin/server

  docker:
    needs: [build]
    runs-on: ubuntu-latest
    if: github.event_name != 'pull_request'
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha,prefix=
            type=raw,value=latest,enable=${{ github.ref == 'refs/heads/main' }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          build-args: |
            VERSION=${{ github.sha }}
```

### Security Scanning Workflow

```yaml
# .github/workflows/security.yml
name: Security

on:
  push:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'CRITICAL,HIGH'
```

---

## 3. Version Embedding

### Build with Version Info

```go
// cmd/api/main.go
package main

import (
    "fmt"
    "runtime"
)

var (
    version   = "dev"
    commit    = "none"
    buildTime = "unknown"
)

func main() {
    fmt.Printf("Version:    %s\n", version)
    fmt.Printf("Commit:     %s\n", commit)
    fmt.Printf("Build Time: %s\n", buildTime)
    fmt.Printf("Go Version: %s\n", runtime.Version())

    // ... rest of main
}
```

### Build Command

```bash
# Build with version info
go build \
    -ldflags="-w -s \
        -X main.version=$(git describe --tags --always) \
        -X main.commit=$(git rev-parse HEAD) \
        -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o ./bin/server ./cmd/api
```

---

## 4. Makefile

```makefile
# Makefile
.PHONY: all build test lint clean docker run

# Variables
APP_NAME := myapp
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -w -s \
    -X main.version=$(VERSION) \
    -X main.commit=$(COMMIT) \
    -X main.buildTime=$(BUILD_TIME)

# Default target
all: lint test build

# Build
build:
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o ./bin/$(APP_NAME) ./cmd/api

# Run
run:
	go run ./cmd/api

# Test
test:
	go test -v -race -cover ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

vet:
	go vet ./...

# Clean
clean:
	rm -rf ./bin
	rm -f coverage.out coverage.html

# Docker
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .

docker-run:
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

# Database
migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down

# Generate
generate:
	go generate ./...

# Dependencies
deps:
	go mod download
	go mod tidy

# Install tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/cosmtrek/air@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter"
	@echo "  docker-build  - Build Docker image"
	@echo "  clean         - Clean build artifacts"
```

---

## 5. Release Notes Template

```markdown
# Release v1.2.0

## Highlights
- Feature X for improved performance
- Bug fix for issue #123

## Breaking Changes
- API endpoint `/v1/users` now requires authentication

## New Features
- Add user search functionality (#45)
- Add export to CSV (#67)

## Bug Fixes
- Fix memory leak in connection pool (#123)
- Fix race condition in cache (#134)

## Performance
- 30% faster response times for list endpoints
- Reduced memory usage by 20%

## Dependencies
- Upgraded Go to 1.21
- Updated chi to v5.0.10

## Migration
```bash
# Run database migrations
make migrate-up

# Update configuration
cp config.example.yaml config.yaml
# Add new required field: cache.ttl
```

## Contributors
- @developer1
- @developer2
```

---

## 6. Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:postgres@db:5432/myapp?sslmode=disable
      - REDIS_URL=redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## 7. .dockerignore

```
# .dockerignore
.git
.gitignore
.github
.vscode
.idea

# Build artifacts
bin/
dist/
*.exe

# Test files
*_test.go
coverage.out
coverage.html

# Documentation
*.md
docs/

# Local config
.env
.env.local
config.local.yaml

# Temporary files
tmp/
*.log
```

---

## Quick Reference

| Task | Command |
|------|---------|
| Build | `make build` |
| Test | `make test` |
| Lint | `make lint` |
| Docker build | `make docker-build` |
| Run locally | `make run` |
| Coverage | `make test-coverage` |

---

## Related Knowledge

- [02-kubernetes.md](./02-kubernetes.md) - K8s deployment
- [../shared/01-go-fundamentals.md](../shared/01-go-fundamentals.md) - Go basics
