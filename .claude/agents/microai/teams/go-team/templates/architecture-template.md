# Architecture Design: {Project Title}

**Date:** {date}
**Author:** Architect Agent
**Status:** Draft | Review | Approved

---

## Overview

{Brief description of the architecture approach}

---

## Architecture Pattern

**Selected Pattern:** {Clean Architecture | Hexagonal | Simple Layered}

**Justification:**
{Why this pattern was chosen for this project}

---

## Package Structure

```
/{project}
├── cmd/
│   └── app/
│       └── main.go              # Entry point, DI wiring
├── internal/
│   ├── handler/                 # HTTP handlers (delivery layer)
│   │   ├── handler.go           # Interface
│   │   └── {feature}_handler.go # Implementation
│   ├── service/                 # Business logic (use case layer)
│   │   ├── service.go           # Interface
│   │   └── {feature}_service.go # Implementation
│   ├── repo/                    # Data access (infrastructure layer)
│   │   ├── repo.go              # Interface
│   │   └── {feature}_repo.go    # Implementation
│   ├── model/                   # Domain models
│   │   ├── {entity}.go
│   │   └── errors.go            # Custom errors
│   └── middleware/              # HTTP middleware
│       ├── logging.go
│       └── auth.go
├── pkg/                         # Public libraries (if needed)
├── configs/
│   └── config.yaml
└── tests/
    └── integration/
```

---

## Interface Definitions

### Handler Layer

```go
// internal/handler/handler.go
type {Feature}Handler interface {
    Get(w http.ResponseWriter, r *http.Request)
    Create(w http.ResponseWriter, r *http.Request)
    Update(w http.ResponseWriter, r *http.Request)
    Delete(w http.ResponseWriter, r *http.Request)
}
```

### Service Layer

```go
// internal/service/service.go
type {Feature}Service interface {
    GetByID(ctx context.Context, id string) (*model.{Entity}, error)
    Create(ctx context.Context, input CreateInput) (*model.{Entity}, error)
    Update(ctx context.Context, id string, input UpdateInput) (*model.{Entity}, error)
    Delete(ctx context.Context, id string) error
}
```

### Repository Layer

```go
// internal/repo/repo.go
type {Feature}Repository interface {
    GetByID(ctx context.Context, id string) (*model.{Entity}, error)
    Create(ctx context.Context, entity *model.{Entity}) error
    Update(ctx context.Context, entity *model.{Entity}) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]*model.{Entity}, error)
}
```

---

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Request                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ HANDLER LAYER                                                │
│ - Validate HTTP input                                        │
│ - Map to domain types                                        │
│ - Call service                                               │
│ - Map response                                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ SERVICE LAYER                                                │
│ - Business logic                                             │
│ - Orchestrate operations                                     │
│ - Call repository                                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ REPOSITORY LAYER                                             │
│ - Data access abstraction                                    │
│ - Database operations                                        │
│ - External service calls                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    DATABASE / EXTERNAL                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Error Strategy

### Custom Error Types

```go
// internal/model/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("resource conflict")
    ErrInternal      = errors.New("internal error")
)
```

### Error Wrapping

```go
// Always wrap errors with context
if err != nil {
    return nil, fmt.Errorf("get user %s: %w", id, err)
}
```

### HTTP Error Mapping

| Domain Error | HTTP Status |
|--------------|-------------|
| ErrNotFound | 404 |
| ErrInvalidInput | 400 |
| ErrUnauthorized | 401 |
| ErrForbidden | 403 |
| ErrConflict | 409 |
| ErrInternal | 500 |

---

## Concurrency Considerations

### Goroutines
- {Where goroutines are used}
- {Synchronization approach}

### Channels
- {Channel usage if any}

### Synchronization
- {Mutex usage}
- {sync.Pool if applicable}

---

## Dependencies

### External Libraries

| Library | Purpose | Version |
|---------|---------|---------|
| github.com/stretchr/testify | Testing | latest |
| {library} | {purpose} | {version} |

### External Services

| Service | Purpose | Connection |
|---------|---------|------------|
| {database} | Data storage | {connection string} |
| {service} | {purpose} | {endpoint} |

---

## Security Considerations

- [ ] Input validation at handler layer
- [ ] Parameterized queries for SQL
- [ ] No sensitive data in logs
- [ ] Proper error messages (no internal details)
- [ ] Context-based timeouts

---

## Diagram

```
[ASCII diagram of component relationships]
```

---

## Approval

| Role | Name | Date | Status |
|------|------|------|--------|
| Observer | | | Pending |
| Coder | | | Pending |
