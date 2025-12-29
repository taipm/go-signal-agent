# Go Fundamentals - Shared Knowledge

**Version:** 1.0.0
**Applies to:** All Agents

---

## TL;DR

- Context là parameter đầu tiên
- Error là return value cuối cùng
- Accept interfaces, return structs
- Wrap errors với `%w`
- Dùng `defer` cho cleanup

---

## 1. Function Signatures

### Context First

```go
// ✅ CORRECT
func GetUser(ctx context.Context, id string) (*User, error)
func CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error)

// ❌ WRONG
func GetUser(id string, ctx context.Context) (*User, error)
func CreateOrder(req CreateOrderRequest) (*Order, error)  // missing context
```

### Error Last

```go
// ✅ CORRECT
func Process(ctx context.Context, data []byte) (Result, error)
func Save(ctx context.Context, item *Item) error

// ❌ WRONG
func Process(ctx context.Context, data []byte) (error, Result)
func Save(ctx context.Context, item *Item) (error, bool)
```

---

## 2. Error Handling

### Always Check Errors

```go
// ✅ CORRECT
result, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}

// ❌ WRONG
result, _ := doSomething()  // ignoring error
doSomething()               // not checking error
```

### Error Wrapping

```go
// ✅ CORRECT - wrap with context
if err != nil {
    return fmt.Errorf("get user %s: %w", userID, err)
}

// ✅ CORRECT - custom error types
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// ❌ WRONG - losing error chain
if err != nil {
    return errors.New("something failed")  // original error lost
}

// ❌ WRONG - no context
if err != nil {
    return err  // no additional context
}
```

### Sentinel Errors

```go
// ✅ CORRECT
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

// Usage
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

---

## 3. Interface Design

### Accept Interfaces, Return Structs

```go
// ✅ CORRECT
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// ❌ WRONG - returning interface
func NewUserService(repo UserRepository) UserServiceInterface {
    return &UserService{repo: repo}
}
```

### Interface Location

```go
// ✅ CORRECT - interface in consumer package
// internal/service/user.go
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

type UserService struct {
    repo UserRepository
}

// ❌ WRONG - interface in provider package
// internal/repository/user.go
type UserRepository interface {  // should be in service package
    GetByID(ctx context.Context, id string) (*User, error)
}
```

### Small Interfaces

```go
// ✅ CORRECT - focused interfaces
type Reader interface {
    Read(ctx context.Context, id string) (*Data, error)
}

type Writer interface {
    Write(ctx context.Context, data *Data) error
}

type ReadWriter interface {
    Reader
    Writer
}

// ❌ WRONG - god interface
type DataManager interface {
    Read(ctx context.Context, id string) (*Data, error)
    Write(ctx context.Context, data *Data) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]*Data, error)
    Count(ctx context.Context) (int, error)
    Search(ctx context.Context, query string) ([]*Data, error)
    // ... 20 more methods
}
```

---

## 4. Resource Management

### Defer for Cleanup

```go
// ✅ CORRECT
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    return io.ReadAll(f)
}

// ✅ CORRECT - defer with error handling
func WriteFile(path string, data []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("close file: %w", cerr)
        }
    }()

    _, err = f.Write(data)
    return err
}

// ❌ WRONG - no cleanup
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    // file never closed!
    return io.ReadAll(f)
}
```

---

## 5. Struct Tags

### JSON Tags

```go
// ✅ CORRECT - snake_case for JSON
type User struct {
    ID        string    `json:"id"`
    FirstName string    `json:"first_name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// ✅ CORRECT - omitempty for optional fields
type UpdateRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
}
```

### Database Tags

```go
// ✅ CORRECT - consistent tag style
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

---

## 6. Naming Conventions

### Variables

```go
// ✅ CORRECT
var (
    userService  *UserService
    httpClient   *http.Client
    ctx          context.Context
    err          error
)

// ❌ WRONG
var (
    user_service *UserService  // no underscores
    HTTPClient   *http.Client  // not exported, use lowercase
    context      context.Context  // shadows package name
)
```

### Constants

```go
// ✅ CORRECT
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
    StatusActive   = "active"
)

// ✅ CORRECT - iota for enums
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusInactive
)
```

### Package Names

```go
// ✅ CORRECT
package user
package repository
package middleware

// ❌ WRONG
package userService  // no camelCase
package user_repo    // no underscores
package utils        // too generic
```

---

## 7. Import Organization

```go
// ✅ CORRECT - grouped imports
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External packages
    "github.com/rs/zerolog"
    "gorm.io/gorm"

    // Internal packages
    "myproject/internal/model"
    "myproject/internal/repository"
)

// ❌ WRONG - unorganized
import (
    "myproject/internal/model"
    "context"
    "github.com/rs/zerolog"
    "fmt"
    "gorm.io/gorm"
)
```

---

## 8. Project Structure

```
project/
├── cmd/
│   └── app/
│       └── main.go           # Entry point, DI wiring
├── internal/
│   ├── handler/              # HTTP handlers
│   ├── service/              # Business logic
│   ├── repository/           # Data access
│   ├── model/                # Domain models
│   └── middleware/           # HTTP middleware
├── pkg/                      # Public libraries
├── configs/                  # Configuration files
├── migrations/               # Database migrations
├── tests/                    # Integration tests
├── go.mod
├── go.sum
└── Makefile
```

---

## Quick Reference Card

| Topic | Pattern |
|-------|---------|
| Context | First parameter |
| Error | Last return value |
| Error wrap | `fmt.Errorf("operation: %w", err)` |
| Interface | Accept interface, return struct |
| Cleanup | `defer resource.Close()` |
| JSON | `json:"snake_case"` |
| Imports | stdlib → external → internal |

---

## Related Knowledge

- [02-error-patterns.md](./02-error-patterns.md) - Chi tiết error handling
- [03-logging-standards.md](./03-logging-standards.md) - Logging với slog/zerolog
- [04-testing-patterns.md](./04-testing-patterns.md) - Testing best practices
