# Architecture Patterns - Architect Agent Knowledge

**Version:** 1.0.0
**Agent:** Architect Agent

---

## TL;DR

- Clean Architecture: domain trung tâm, không phụ thuộc infrastructure
- Dependency Rule: outer layers depend on inner layers
- Accept interfaces, return structs
- Interface định nghĩa ở package consumer

---

## 1. Clean Architecture

### Layer Structure

```
┌─────────────────────────────────────────────────────────────┐
│                      External Systems                        │
│              (Database, APIs, UI, Frameworks)                │
├─────────────────────────────────────────────────────────────┤
│                    Interface Adapters                        │
│            (Controllers, Gateways, Presenters)               │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                         │
│              (Use Cases, Application Services)               │
├─────────────────────────────────────────────────────────────┤
│                      Domain Layer                            │
│            (Entities, Value Objects, Domain Services)        │
└─────────────────────────────────────────────────────────────┘
```

### Go Project Structure

```
myproject/
├── cmd/
│   └── api/
│       └── main.go              # Wiring, DI
│
├── internal/
│   ├── domain/                  # Domain Layer
│   │   ├── user.go              # Entity
│   │   ├── order.go             # Entity
│   │   └── errors.go            # Domain errors
│   │
│   ├── usecase/                 # Application Layer
│   │   ├── user/
│   │   │   ├── service.go       # Use case implementation
│   │   │   └── interface.go     # Port definitions
│   │   └── order/
│   │       ├── service.go
│   │       └── interface.go
│   │
│   ├── adapter/                 # Interface Adapters
│   │   ├── http/                # HTTP handlers
│   │   │   ├── handler.go
│   │   │   └── middleware.go
│   │   └── repository/          # Data access
│   │       ├── postgres/
│   │       │   ├── user.go
│   │       │   └── order.go
│   │       └── redis/
│   │           └── cache.go
│   │
│   └── infrastructure/          # External concerns
│       ├── config/
│       ├── database/
│       └── logger/
│
├── pkg/                         # Public utilities
├── configs/
└── migrations/
```

### Dependency Direction

```go
// ✅ CORRECT - inner layers don't know about outer
// domain/user.go
type User struct {
    ID    string
    Email string
    Name  string
}

// usecase/user/interface.go - defines what it needs
type UserRepository interface {
    Save(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
}

// usecase/user/service.go - implements business logic
type UserService struct {
    repo UserRepository  // depends on interface, not implementation
}

func (s *UserService) CreateUser(ctx context.Context, email, name string) (*domain.User, error) {
    user := &domain.User{
        ID:    uuid.New().String(),
        Email: email,
        Name:  name,
    }
    if err := s.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("save user: %w", err)
    }
    return user, nil
}

// adapter/repository/postgres/user.go - implements interface
type PostgresUserRepo struct {
    db *sql.DB
}

func (r *PostgresUserRepo) Save(ctx context.Context, user *domain.User) error {
    _, err := r.db.ExecContext(ctx,
        "INSERT INTO users (id, email, name) VALUES ($1, $2, $3)",
        user.ID, user.Email, user.Name,
    )
    return err
}

// ❌ WRONG - domain depending on infrastructure
// domain/user.go
import "gorm.io/gorm"  // domain shouldn't know about GORM

type User struct {
    gorm.Model  // infrastructure leak!
    Email string
}
```

---

## 2. Hexagonal Architecture (Ports & Adapters)

### Structure

```
                    ┌─────────────────┐
                    │   HTTP Adapter  │
                    └────────┬────────┘
                             │
         ┌───────────────────┴───────────────────┐
         │                                       │
         │           APPLICATION CORE            │
         │                                       │
         │  ┌─────────┐         ┌─────────┐     │
         │  │  Ports  │◄───────►│ Domain  │     │
         │  │  (API)  │         │ (Logic) │     │
         │  └─────────┘         └─────────┘     │
         │                                       │
         └───────────────────┬───────────────────┘
                             │
                    ┌────────┴────────┐
                    │   DB Adapter    │
                    └─────────────────┘
```

### Go Implementation

```go
// ports/input.go - Primary/Driving Ports
type UserService interface {
    CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
    GetUser(ctx context.Context, id string) (*User, error)
}

// ports/output.go - Secondary/Driven Ports
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
}

type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}

// adapters/primary/http.go - Primary Adapter
type HTTPHandler struct {
    userService ports.UserService  // uses port
}

func (h *HTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // parse request, call service, return response
}

// adapters/secondary/postgres.go - Secondary Adapter
type PostgresRepository struct {
    db *sql.DB
}

// implements ports.UserRepository
```

---

## 3. Simple Layered Architecture

### For Smaller Projects

```
myproject/
├── cmd/api/main.go
├── internal/
│   ├── handler/        # HTTP layer
│   │   └── user.go
│   ├── service/        # Business logic
│   │   └── user.go
│   ├── repository/     # Data access
│   │   └── user.go
│   └── model/          # Data structures
│       └── user.go
├── pkg/
└── configs/
```

### Implementation

```go
// model/user.go
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// repository/user.go
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id string) (*model.User, error)
}

type userRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &userRepository{db: db}
}

// service/user.go
type UserService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, email, name string) (*model.User, error) {
    user := &model.User{
        ID:        uuid.New().String(),
        Email:     email,
        Name:      name,
        CreatedAt: time.Now(),
    }
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }
    return user, nil
}

// handler/user.go
type UserHandler struct {
    service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
    return &UserHandler{service: service}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    // parse request, call service, return response
}
```

---

## 4. Dependency Injection

### Constructor Injection

```go
// ✅ CORRECT - inject dependencies via constructor
type UserService struct {
    repo   UserRepository
    cache  Cache
    logger *slog.Logger
}

func NewUserService(repo UserRepository, cache Cache, logger *slog.Logger) *UserService {
    return &UserService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}

// main.go - wire everything
func main() {
    // Create infrastructure
    db := setupDatabase()
    redis := setupRedis()
    logger := setupLogger()

    // Create repositories
    userRepo := repository.NewUserRepository(db)
    cache := cache.NewRedisCache(redis)

    // Create services
    userService := service.NewUserService(userRepo, cache, logger)

    // Create handlers
    userHandler := handler.NewUserHandler(userService)

    // Setup routes
    router := chi.NewRouter()
    router.Post("/users", userHandler.Create)
    router.Get("/users/{id}", userHandler.Get)

    http.ListenAndServe(":8080", router)
}
```

### Interface Definition Location

```go
// ✅ CORRECT - interface in consumer package
// service/user.go
type UserRepository interface {  // defined where it's used
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
}

type UserService struct {
    repo UserRepository
}

// ❌ WRONG - interface in provider package
// repository/user.go
type UserRepository interface {  // shouldn't be here
    Create(ctx context.Context, user *User) error
}

type userRepository struct { /* ... */ }
```

---

## 5. Package Design

### Package Guidelines

```go
// ✅ GOOD - cohesive package
package user

type User struct { /* ... */ }
type CreateRequest struct { /* ... */ }
type Service struct { /* ... */ }
func NewService() *Service { /* ... */ }

// ❌ BAD - utility grab bag
package utils

func FormatDate() { /* ... */ }
func ValidateEmail() { /* ... */ }
func HashPassword() { /* ... */ }
func GenerateID() { /* ... */ }
```

### Avoid Circular Dependencies

```go
// ❌ WRONG - circular dependency
// package user imports package order
// package order imports package user

// ✅ CORRECT - extract shared types
// package domain - shared entities
type User struct { /* ... */ }
type Order struct { /* ... */ }

// package user - uses domain.User
// package order - uses domain.User, domain.Order
```

---

## 6. Architecture Decision Record (ADR)

### Template

```markdown
# ADR-001: Use Clean Architecture

## Status
Accepted

## Context
We need to structure our codebase for maintainability and testability.

## Decision
We will use Clean Architecture with the following layers:
- Domain (entities, value objects)
- Use Cases (application services)
- Adapters (HTTP handlers, repositories)
- Infrastructure (database, external services)

## Consequences
**Positive:**
- Clear separation of concerns
- Easy to test (mock interfaces)
- Business logic isolated from frameworks

**Negative:**
- More boilerplate code
- Learning curve for new developers
- More directories and files
```

---

## Quick Reference

| Pattern | When to Use |
|---------|-------------|
| Clean Architecture | Large, complex domains |
| Hexagonal | Many external integrations |
| Simple Layered | Small to medium projects |

| Principle | Rule |
|-----------|------|
| Dependency | Outer → Inner |
| Interface Location | Consumer package |
| DI | Constructor injection |

---

## Related Knowledge

- [02-api-design.md](./02-api-design.md) - REST API patterns
- [../shared/01-go-fundamentals.md](../shared/01-go-fundamentals.md) - Go basics
