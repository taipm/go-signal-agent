---
stepNumber: 4
nextStep: './step-05-testing.md'
agent: go-coder-agent
hasBreakpoint: false
---

# Step 04: Code Implementation

## STEP GOAL

Go Coder Agent implements the code based on the architecture design, following Go idioms and best practices.

## AGENT ACTIVATION

Load persona từ `../agents/go-coder-agent.md`

Input context:
- Spec từ step 02
- Architecture từ step 03
- Existing code (if any)

## EXECUTION SEQUENCE

### 1. Coder Agent Introduction

```
[Go Coder Agent]

Đã nhận architecture design. Bắt đầu implementation cho "{topic}".

Implementation order:
1. Models/types
2. Interfaces
3. Repository
4. Service
5. Handler
6. Main (DI wiring)
```

### 2. Create Folder Structure

```bash
mkdir -p cmd/app internal/{handler,service,repo,model,middleware} configs tests/integration
```

### 3. Implement Models

```go
// internal/model/{entity}.go
package model

import "time"

type {Entity} struct {
    ID        string    `json:"id"`
    // ... fields based on spec
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// internal/model/errors.go
package model

import "errors"

var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)
```

### 4. Implement Repository Interface

```go
// internal/repo/repo.go
package repo

import (
    "context"
    "{module}/internal/model"
)

type {Entity}Repository interface {
    GetByID(ctx context.Context, id string) (*model.{Entity}, error)
    Create(ctx context.Context, entity *model.{Entity}) error
    Update(ctx context.Context, entity *model.{Entity}) error
    Delete(ctx context.Context, id string) error
}
```

### 5. Implement Repository (In-Memory hoặc Real)

```go
// internal/repo/memory_{entity}_repo.go
package repo

import (
    "context"
    "sync"
    "{module}/internal/model"
)

type Memory{Entity}Repo struct {
    mu   sync.RWMutex
    data map[string]*model.{Entity}
}

func NewMemory{Entity}Repo() *Memory{Entity}Repo {
    return &Memory{Entity}Repo{
        data: make(map[string]*model.{Entity}),
    }
}

func (r *Memory{Entity}Repo) GetByID(ctx context.Context, id string) (*model.{Entity}, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    entity, ok := r.data[id]
    if !ok {
        return nil, model.ErrNotFound
    }
    return entity, nil
}
```

### 6. Implement Service

```go
// internal/service/{entity}_service.go
package service

import (
    "context"
    "fmt"
    "log/slog"
    "{module}/internal/model"
    "{module}/internal/repo"
)

type {Entity}Service struct {
    repo repo.{Entity}Repository
    log  *slog.Logger
}

func New{Entity}Service(repo repo.{Entity}Repository, log *slog.Logger) *{Entity}Service {
    return &{Entity}Service{repo: repo, log: log}
}

func (s *{Entity}Service) GetByID(ctx context.Context, id string) (*model.{Entity}, error) {
    entity, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get {entity} by id %s: %w", id, err)
    }
    return entity, nil
}
```

### 7. Implement Handler

```go
// internal/handler/{entity}_handler.go
package handler

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "{module}/internal/service"
)

type {Entity}Handler struct {
    svc *service.{Entity}Service
    log *slog.Logger
}

func New{Entity}Handler(svc *service.{Entity}Service, log *slog.Logger) *{Entity}Handler {
    return &{Entity}Handler{svc: svc, log: log}
}

func (h *{Entity}Handler) Get(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := r.PathValue("id")

    if id == "" {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }

    entity, err := h.svc.GetByID(ctx, id)
    if err != nil {
        h.log.Error("failed to get entity", "error", err, "id", id)
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(entity)
}
```

### 8. Implement Main with DI

```go
// cmd/app/main.go
package main

import (
    "log/slog"
    "net/http"
    "os"
    "{module}/internal/handler"
    "{module}/internal/repo"
    "{module}/internal/service"
)

func main() {
    log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // DI wiring
    repository := repo.NewMemory{Entity}Repo()
    svc := service.New{Entity}Service(repository, log)
    h := handler.New{Entity}Handler(svc, log)

    // Routes
    mux := http.NewServeMux()
    mux.HandleFunc("GET /{entities}/{id}", h.Get)
    mux.HandleFunc("POST /{entities}", h.Create)

    log.Info("starting server", "addr", ":8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Error("server error", "error", err)
        os.Exit(1)
    }
}
```

### 9. Verify Build

```bash
go mod tidy
go build ./...
go vet ./...
```

### 10. Report Progress

```
[Go Coder Agent]

Implementation complete:

Files created:
- cmd/app/main.go
- internal/model/{entity}.go
- internal/model/errors.go
- internal/repo/repo.go
- internal/repo/memory_{entity}_repo.go
- internal/service/{entity}_service.go
- internal/handler/{entity}_handler.go

Build status: PASS

Ready for Testing phase.
```

## OUTPUT

```yaml
outputs:
  code_files:
    - path: "cmd/app/main.go"
      purpose: "Entry point"
    - path: "internal/model/{entity}.go"
      purpose: "Domain model"
    # ... more files
  build_status: "pass"
```

## SUCCESS CRITERIA

- [ ] All planned files created
- [ ] Code follows architecture design
- [ ] Build passes (`go build ./...`)
- [ ] No obvious errors from `go vet`
- [ ] Ready for Testing phase

## NEXT STEP

Load `./step-05-testing.md`
