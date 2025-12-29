# Implementation Order - Coder Agent Knowledge

**Version:** 1.0.0
**Agent:** Go Coder Agent

---

## TL;DR

- Models/Types trước
- Interfaces tiếp theo
- Repository → Service → Handler
- main.go cuối cùng (DI wiring)
- Mỗi layer test được độc lập

---

## 1. Implementation Sequence

### Standard Order

```
1. models/types     → Data structures
2. errors           → Custom error types
3. interfaces       → Contracts between layers
4. repository       → Data access implementation
5. service          → Business logic
6. handler          → HTTP layer
7. middleware       → Cross-cutting concerns
8. main.go          → DI wiring and bootstrap
```

### Why This Order?

```go
// Step 1: Models - everything depends on these
type User struct {
    ID    string
    Email string
    Name  string
}

// Step 2: Errors - service needs to return these
var ErrUserNotFound = errors.New("user not found")

// Step 3: Interfaces - contracts before implementation
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
}

// Step 4: Repository - implements interface
type userRepo struct { db *sql.DB }

func (r *userRepo) FindByID(ctx context.Context, id string) (*User, error) {
    // implementation
}

// Step 5: Service - uses repository via interface
type UserService struct {
    repo UserRepository  // interface, not concrete
}

// Step 6: Handler - uses service
type UserHandler struct {
    service *UserService
}

// Step 7: Middleware
func AuthMiddleware(next http.Handler) http.Handler { ... }

// Step 8: main.go - wire everything
func main() {
    db := setupDB()
    repo := NewUserRepo(db)
    service := NewUserService(repo)
    handler := NewUserHandler(service)
    // setup routes
}
```

---

## 2. Model Implementation

### Entity Definition

```go
// internal/model/user.go

package model

import "time"

// User represents a user in the system
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Name      string    `json:"name" db:"name"`
    Password  string    `json:"-" db:"password"`  // never expose
    Status    Status    `json:"status" db:"status"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Status enum
type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusPending  Status = "pending"
)
```

### Request/Response DTOs

```go
// internal/model/dto.go

// CreateUserRequest is the input for creating a user
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8"`
}

// CreateUserResponse is returned after creating a user
type CreateUserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

// FromUser creates response from user entity
func (r *CreateUserResponse) FromUser(u *User) {
    r.ID = u.ID
    r.Email = u.Email
    r.Name = u.Name
    r.CreatedAt = u.CreatedAt
}
```

---

## 3. Error Definition

```go
// internal/model/errors.go

package model

import "errors"

// Domain errors
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
)

// Specific errors
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyUsed  = errors.New("email already in use")
    ErrInvalidCredentials = errors.New("invalid credentials")
)

// ValidationError contains field-level errors
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

---

## 4. Interface Definition

```go
// internal/service/interface.go

package service

import (
    "context"
    "myproject/internal/model"
)

// UserRepository defines data access operations
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id string) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*model.User, error)
}

// Cache defines caching operations
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

// EmailSender defines email operations
type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}
```

---

## 5. Repository Implementation

```go
// internal/repository/user.go

package repository

import (
    "context"
    "database/sql"
    "fmt"

    "myproject/internal/model"
)

type userRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    query := `
        INSERT INTO users (id, email, name, password, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    _, err := r.db.ExecContext(ctx, query,
        user.ID, user.Email, user.Name, user.Password,
        user.Status, user.CreatedAt, user.UpdatedAt,
    )
    if err != nil {
        return fmt.Errorf("insert user: %w", err)
    }
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    query := `SELECT id, email, name, status, created_at, updated_at FROM users WHERE id = $1`

    user := &model.User{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.Name,
        &user.Status, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, model.ErrUserNotFound
        }
        return nil, fmt.Errorf("query user: %w", err)
    }
    return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    query := `SELECT id, email, name, password, status, created_at, updated_at FROM users WHERE email = $1`

    user := &model.User{}
    err := r.db.QueryRowContext(ctx, query, email).Scan(
        &user.ID, &user.Email, &user.Name, &user.Password,
        &user.Status, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, model.ErrUserNotFound
        }
        return nil, fmt.Errorf("query user by email: %w", err)
    }
    return user, nil
}
```

---

## 6. Service Implementation

```go
// internal/service/user.go

package service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"

    "myproject/internal/model"
)

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

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
    // Check if email exists
    existing, err := s.repo.GetByEmail(ctx, req.Email)
    if err != nil && !errors.Is(err, model.ErrUserNotFound) {
        return nil, fmt.Errorf("check existing user: %w", err)
    }
    if existing != nil {
        return nil, model.ErrEmailAlreadyUsed
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }

    // Create user
    now := time.Now()
    user := &model.User{
        ID:        uuid.New().String(),
        Email:     req.Email,
        Name:      req.Name,
        Password:  string(hashedPassword),
        Status:    model.StatusActive,
        CreatedAt: now,
        UpdatedAt: now,
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    s.logger.Info("user created",
        slog.String("user_id", user.ID),
        slog.String("email", user.Email),
    )

    return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*model.User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    return user, nil
}
```

---

## 7. Handler Implementation

```go
// internal/handler/user.go

package handler

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/go-chi/chi/v5"

    "myproject/internal/model"
    "myproject/internal/service"
)

type UserHandler struct {
    service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
    return &UserHandler{service: service}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req model.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    user, err := h.service.CreateUser(r.Context(), req)
    if err != nil {
        if errors.Is(err, model.ErrEmailAlreadyUsed) {
            respondError(w, http.StatusConflict, "email already in use")
            return
        }
        respondError(w, http.StatusInternalServerError, "failed to create user")
        return
    }

    response := &model.CreateUserResponse{}
    response.FromUser(user)

    respondJSON(w, http.StatusCreated, response)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        if errors.Is(err, model.ErrUserNotFound) {
            respondError(w, http.StatusNotFound, "user not found")
            return
        }
        respondError(w, http.StatusInternalServerError, "failed to get user")
        return
    }

    respondJSON(w, http.StatusOK, user)
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}
```

---

## 8. Main.go Wiring

```go
// cmd/api/main.go

package main

import (
    "database/sql"
    "log/slog"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    _ "github.com/lib/pq"

    "myproject/internal/handler"
    "myproject/internal/repository"
    "myproject/internal/service"
)

func main() {
    // Setup logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Setup database
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        logger.Error("failed to connect to database", slog.Any("error", err))
        os.Exit(1)
    }
    defer db.Close()

    // Create repositories
    userRepo := repository.NewUserRepository(db)

    // Create services
    userService := service.NewUserService(userRepo, nil, logger)

    // Create handlers
    userHandler := handler.NewUserHandler(userService)

    // Setup router
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)

    // Routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/users", userHandler.Create)
        r.Get("/users/{id}", userHandler.Get)
    })

    // Start server
    logger.Info("server starting", slog.String("addr", ":8080"))
    if err := http.ListenAndServe(":8080", r); err != nil {
        logger.Error("server failed", slog.Any("error", err))
        os.Exit(1)
    }
}
```

---

## Quick Reference

| Order | Layer | Package | Depends On |
|-------|-------|---------|------------|
| 1 | Models | `model/` | - |
| 2 | Errors | `model/` | - |
| 3 | Interfaces | `service/` | model |
| 4 | Repository | `repository/` | model |
| 5 | Service | `service/` | model, interface |
| 6 | Handler | `handler/` | model, service |
| 7 | Main | `cmd/` | all |

---

## Related Knowledge

- [02-error-handling.md](./02-error-handling.md) - Error patterns
- [../shared/01-go-fundamentals.md](../shared/01-go-fundamentals.md) - Go basics
