---
stepNumber: 3
nextStep: './step-04-implementation.md'
agent: architect-agent
hasBreakpoint: true
---

# Step 03: Architecture Design

## STEP GOAL

Architect Agent thiết kế system architecture dựa trên spec từ PM Agent, chọn patterns phù hợp, và định nghĩa package structure.

## AGENT ACTIVATION

Load persona từ `../agents/architect-agent.md`

Input context:
- Spec từ step 02
- Project context (existing code structure if any)
- Go best practices

## EXECUTION SEQUENCE

### 1. Architect Agent Introduction

```
[Architect Agent]

Đã nhận spec từ PM Agent. Tôi sẽ thiết kế architecture cho "{topic}".

Analyzing requirements...
```

### 2. Pattern Selection

Architect quyết định:
- **Architecture Pattern:** Clean Architecture / Hexagonal / Simple
- **Data Access:** Repository pattern
- **Business Logic:** Service pattern
- **API Layer:** Handler pattern

Giải thích lý do chọn pattern.

### 3. Package Structure Design

```
/{project}
├── cmd/
│   └── app/
│       └── main.go              # Entry point, DI wiring
├── internal/
│   ├── handler/                 # HTTP handlers
│   │   ├── handler.go           # Interface
│   │   └── {feature}_handler.go # Implementation
│   ├── service/                 # Business logic
│   │   ├── service.go           # Interface
│   │   └── {feature}_service.go # Implementation
│   ├── repo/                    # Data access
│   │   ├── repo.go              # Interface
│   │   └── {feature}_repo.go    # Implementation
│   ├── model/                   # Domain models
│   │   ├── {entity}.go
│   │   └── errors.go            # Custom errors
│   └── middleware/              # HTTP middleware
│       └── logging.go
├── pkg/                         # Public libraries (if needed)
├── configs/
│   └── config.yaml
└── tests/
    └── integration/
```

### 4. Interface Definitions

```go
// handler/handler.go
type {Feature}Handler interface {
    Get(w http.ResponseWriter, r *http.Request)
    Create(w http.ResponseWriter, r *http.Request)
    // ...
}

// service/service.go
type {Feature}Service interface {
    GetByID(ctx context.Context, id string) (*model.{Entity}, error)
    Create(ctx context.Context, input CreateInput) (*model.{Entity}, error)
    // ...
}

// repo/repo.go
type {Feature}Repository interface {
    GetByID(ctx context.Context, id string) (*model.{Entity}, error)
    Create(ctx context.Context, entity *model.{Entity}) error
    // ...
}
```

### 5. Data Flow Diagram

```
Request
   ↓
[Handler] ← validates input, maps to domain
   ↓
[Service] ← business logic, orchestrates
   ↓
[Repository] ← data access
   ↓
[Database/External]
```

### 6. Error Strategy

```go
// model/errors.go
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUnauthorized = errors.New("unauthorized")
)

// Usage: fmt.Errorf("get user %s: %w", id, ErrNotFound)
```

### 7. Present Design to Observer

```
[Architect Agent]

Architecture design cho "{topic}":

{formatted design document}

---
═══════════════ BREAKPOINT ═══════════════

Observer, vui lòng review architecture này.

Options:
- [Enter] → Approve và tiếp tục đến Implementation
- @arch: <feedback> → Yêu cầu Architect điều chỉnh
- *pause → Tạm dừng để review
```

## OUTPUT

Architecture document saved to session state:
```yaml
outputs:
  architecture:
    pattern: "clean-architecture"
    packages: [...]
    interfaces: [...]
    data_flow: "..."
    error_strategy: "..."
    approved: true/false
```

## SUCCESS CRITERIA

- [ ] Architecture pattern selected and justified
- [ ] Package structure defined
- [ ] Interfaces documented
- [ ] Data flow clear
- [ ] Observer approved design
- [ ] Ready for Implementation phase

## NEXT STEP

After breakpoint approval, load `./step-04-implementation.md`
