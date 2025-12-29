# Error Patterns - Shared Knowledge

**Version:** 1.0.0
**Applies to:** All Agents

---

## TL;DR

- Luôn wrap errors với context: `fmt.Errorf("operation: %w", err)`
- Dùng sentinel errors cho known conditions
- Custom error types cho rich error info
- Check errors với `errors.Is()` và `errors.As()`
- Không log và return cùng lúc

---

## 1. Error Wrapping Patterns

### Basic Wrapping

```go
// ✅ CORRECT - add context at each level
func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := r.db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    return user, nil
}

// Result: "get user: query user abc123: sql: no rows"
```

### Error Message Style

```go
// ✅ CORRECT
fmt.Errorf("get user by id: %w", err)      // lowercase, no period
fmt.Errorf("parse config file: %w", err)   // action description
fmt.Errorf("connect to database: %w", err) // verb phrase

// ❌ WRONG
fmt.Errorf("Error getting user: %w", err)  // don't prefix with "Error"
fmt.Errorf("failed to get user: %w", err)  // redundant "failed to"
fmt.Errorf("GetUser: %w", err)             // don't use function name
fmt.Errorf("get user.: %w", err)           // no period
```

---

## 2. Sentinel Errors

### Definition

```go
// ✅ CORRECT - package-level sentinel errors
package domain

import "errors"

var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrInvalidInput  = errors.New("invalid input")
    ErrConflict      = errors.New("conflict")
)
```

### Usage

```go
// ✅ CORRECT - wrap sentinel error with context
func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := r.db.Get(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("user %s: %w", id, domain.ErrNotFound)
        }
        return nil, fmt.Errorf("query user: %w", err)
    }
    return user, nil
}

// ✅ CORRECT - check sentinel error
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

---

## 3. Custom Error Types

### Structured Error

```go
// ✅ CORRECT - rich error information
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// Usage
func ValidateUser(user *User) error {
    if user.Email == "" {
        return &ValidationError{
            Field:   "email",
            Value:   user.Email,
            Message: "email is required",
        }
    }
    return nil
}

// Check with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    log.Printf("Field %s has invalid value: %v", valErr.Field, valErr.Value)
}
```

### Error with Code

```go
// ✅ CORRECT - error with code for API responses
type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Usage
func NewNotFoundError(resource string, id string) *AppError {
    return &AppError{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s with id %s not found", resource, id),
    }
}
```

---

## 4. Error Checking

### errors.Is vs errors.As

```go
// ✅ errors.Is - check specific error value
if errors.Is(err, domain.ErrNotFound) {
    // handle not found
}

if errors.Is(err, context.Canceled) {
    // handle cancellation
}

if errors.Is(err, context.DeadlineExceeded) {
    // handle timeout
}

// ✅ errors.As - extract error type
var valErr *ValidationError
if errors.As(err, &valErr) {
    // use valErr.Field, valErr.Message
}

var appErr *AppError
if errors.As(err, &appErr) {
    // use appErr.Code, appErr.Message
}
```

### Multiple Error Types

```go
// ✅ CORRECT - handle multiple error types
func HandleError(err error) {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        // 404
    case errors.Is(err, domain.ErrUnauthorized):
        // 401
    case errors.Is(err, domain.ErrForbidden):
        // 403
    default:
        var valErr *ValidationError
        if errors.As(err, &valErr) {
            // 400 with validation details
            return
        }
        // 500
    }
}
```

---

## 5. Error Anti-Patterns

### Don't Log and Return

```go
// ❌ WRONG - log and return (error logged twice)
func Process(ctx context.Context) error {
    err := doSomething()
    if err != nil {
        log.Error().Err(err).Msg("failed to do something")
        return err  // caller will also log
    }
    return nil
}

// ✅ CORRECT - only return, let caller decide
func Process(ctx context.Context) error {
    err := doSomething()
    if err != nil {
        return fmt.Errorf("do something: %w", err)
    }
    return nil
}

// Log at the top level (handler)
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    err := h.service.Process(r.Context())
    if err != nil {
        log.Error().Err(err).Msg("process failed")
        http.Error(w, "Internal error", 500)
        return
    }
}
```

### Don't Ignore Errors

```go
// ❌ WRONG
result, _ := json.Marshal(data)  // ignoring marshal error
http.ListenAndServe(":8080", nil)  // ignoring error

// ✅ CORRECT
result, err := json.Marshal(data)
if err != nil {
    return fmt.Errorf("marshal data: %w", err)
}

if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal().Err(err).Msg("server failed")
}
```

### Don't Create New Error Losing Chain

```go
// ❌ WRONG - losing error chain
if err != nil {
    return errors.New("operation failed")  // original err lost
}

// ❌ WRONG - using %v instead of %w
if err != nil {
    return fmt.Errorf("operation failed: %v", err)  // can't unwrap
}

// ✅ CORRECT
if err != nil {
    return fmt.Errorf("operation failed: %w", err)  // preserves chain
}
```

---

## 6. Context Errors

```go
// ✅ CORRECT - check context errors
func LongOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()  // context.Canceled or context.DeadlineExceeded
        default:
            // do work
        }
    }
}

// ✅ CORRECT - wrap context error
func Process(ctx context.Context, items []Item) error {
    for _, item := range items {
        if err := ctx.Err(); err != nil {
            return fmt.Errorf("process items: %w", err)
        }
        if err := processItem(ctx, item); err != nil {
            return fmt.Errorf("process item %s: %w", item.ID, err)
        }
    }
    return nil
}
```

---

## 7. Error Response Mapping

```go
// ✅ CORRECT - map domain errors to HTTP status
func ErrorToStatus(err error) int {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, domain.ErrUnauthorized):
        return http.StatusUnauthorized
    case errors.Is(err, domain.ErrForbidden):
        return http.StatusForbidden
    case errors.Is(err, domain.ErrInvalidInput):
        return http.StatusBadRequest
    case errors.Is(err, domain.ErrConflict):
        return http.StatusConflict
    case errors.Is(err, context.DeadlineExceeded):
        return http.StatusGatewayTimeout
    default:
        return http.StatusInternalServerError
    }
}

// Usage in handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        status := ErrorToStatus(err)
        http.Error(w, err.Error(), status)
        return
    }
    // ...
}
```

---

## Quick Reference

| Pattern | Example |
|---------|---------|
| Wrap error | `fmt.Errorf("operation: %w", err)` |
| Sentinel | `var ErrNotFound = errors.New("not found")` |
| Check value | `errors.Is(err, ErrNotFound)` |
| Check type | `errors.As(err, &customErr)` |
| Context err | `ctx.Err()` returns `Canceled` or `DeadlineExceeded` |

---

## Related Knowledge

- [01-go-fundamentals.md](./01-go-fundamentals.md) - Basic Go patterns
- [03-logging-standards.md](./03-logging-standards.md) - How to log errors
