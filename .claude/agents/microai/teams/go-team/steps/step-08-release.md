---
stepNumber: 8
nextStep: './step-09-synthesis.md'
agent: devops-agent
hasBreakpoint: false
---

# Step 08: Release Configuration

## STEP GOAL

DevOps Agent creates deployment artifacts including Dockerfile, CI/CD pipeline, and release configuration.

## AGENT ACTIVATION

Load persona từ `../agents/devops-agent.md`

Input context:
- Project structure từ implementation
- Build requirements
- Deployment target (if specified)

## EXECUTION SEQUENCE

### 1. DevOps Agent Introduction

```
[DevOps Agent]

Creating release configuration cho "{topic}"...

Artifacts to create:
1. Dockerfile (multi-stage)
2. GitHub Actions CI/CD
3. Makefile
4. .gitignore (if needed)
```

### 2. Create Dockerfile

```dockerfile
# Dockerfile
# Multi-stage build for minimal image

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION:-dev}" \
    -o server ./cmd/app

# Final stage
FROM scratch

# Copy binary
COPY --from=builder /app/server /server

# Copy CA certs for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Non-root user
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### 3. Create GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

      - name: Test
        run: go test -race -coverprofile=coverage.out ./...

      - name: Coverage
        run: go tool cover -func=coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Build
        run: |
          CGO_ENABLED=0 go build \
            -ldflags="-w -s -X main.version=${{ github.sha }}" \
            -o bin/server ./cmd/app

      - uses: actions/upload-artifact@v4
        with:
          name: server
          path: bin/server

  docker:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-buildx-action@v3

      - uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: app:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### 4. Create Makefile

```makefile
# Makefile
.PHONY: build test lint docker clean run

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-w -s -X main.version=$(VERSION)"

## Build the application
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/server ./cmd/app

## Run all tests
test:
	go test -race -cover ./...

## Run linter
lint:
	golangci-lint run

## Build docker image
docker:
	docker build -t app:$(VERSION) .

## Clean build artifacts
clean:
	rm -rf bin/

## Run the application
run: build
	./bin/server

## Show help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
```

### 5. Create/Update .gitignore

```
# .gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test
*.test
coverage.out
coverage.html

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
*.env

# Build
vendor/
```

### 6. Report Results

```
[DevOps Agent]

Release configuration complete:

### Files Created

1. **Dockerfile**
   - Multi-stage build
   - Base: scratch (~10MB final image)
   - Non-root user

2. **.github/workflows/ci.yml**
   - Triggers: push/PR to main
   - Jobs: lint → test → build → docker
   - Caching enabled

3. **Makefile**
   - build, test, lint, docker, clean, run

4. **.gitignore**
   - Standard Go ignores

### Quick Start

```bash
# Local development
make build
make run

# Run tests
make test

# Build docker
make docker

# Or directly
docker build -t app .
docker run -p 8080:8080 app
```

Ready for final synthesis.
```

## OUTPUT

```yaml
outputs:
  release_config:
    dockerfile: "Dockerfile"
    ci_pipeline: ".github/workflows/ci.yml"
    makefile: "Makefile"
    gitignore: ".gitignore"
  docker:
    base: "scratch"
    size: "~10MB"
    user: "non-root"
  ci:
    triggers: ["push", "pull_request"]
    jobs: ["lint", "test", "build", "docker"]
```

## SUCCESS CRITERIA

- [ ] Dockerfile created (multi-stage)
- [ ] CI/CD pipeline created
- [ ] Makefile with common targets
- [ ] .gitignore updated
- [ ] Ready for final synthesis

## NEXT STEP

Load `./step-09-synthesis.md`
