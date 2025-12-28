# Go Idioms and Best Practices

## Error Handling

### The Cardinal Rule
```go
// BROKEN - Never do this
result, _ := someFunction()

// CORRECT - Always check errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("someFunction failed: %w", err)
}
```

### Error Wrapping
```go
// SMELL - No context
if err != nil {
    return err
}

// OK - Adds context
if err != nil {
    return fmt.Errorf("failed to process user %d: %w", userID, err)
}
```

### Sentinel Errors
```go
// Define at package level
var ErrNotFound = errors.New("not found")

// Check with errors.Is
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

### Custom Error Types
```go
// For errors needing additional context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Check with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    // handle validation error
}
```

## Naming Conventions

### Variables
```go
// SMELL - Too verbose
userIdentifier := getUserID()
temporaryBuffer := make([]byte, 1024)

// OK - Short and clear
userID := getUserID()
buf := make([]byte, 1024)
```

### No Stuttering
```go
// SMELL - Package name repeated
package user
type UserService struct{} // user.UserService stutters

// OK - Clean
package user
type Service struct{} // user.Service is clean
```

### Acronyms
```go
// SMELL - Inconsistent
userId, httpUrl, xmlParser

// OK - All caps for acronyms
userID, httpURL, xmlParser
// or if at start
ID, URL, XML
```

### Interfaces
```go
// Single method interface: method name + "er"
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }

// Multi-method: descriptive noun
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

## Package Design

### Single Responsibility
```go
// SMELL - Package does too much
package utils // contains logging, http, database, strings...

// OK - Focused packages
package logger
package httpclient
package database
```

### Minimal Public API
```go
// SMELL - Everything exported
type UserService struct {
    DB        *sql.DB      // Why is this public?
    Cache     *redis.Client
    HTTPClient *http.Client
}

// OK - Only expose what's needed
type UserService struct {
    db        *sql.DB
    cache     *redis.Client
    client    *http.Client
}

func NewUserService(db *sql.DB, cache *redis.Client) *UserService {
    return &UserService{db: db, cache: cache, client: http.DefaultClient}
}
```

### No Circular Dependencies
```go
// BROKEN - Creates import cycle
// package a imports package b
// package b imports package a

// OK - Use interfaces to break cycles
// package a defines interface
// package b implements interface
// package c wires them together
```

## Functions

### Accept Interfaces, Return Structs
```go
// OK - Flexible input, concrete output
func NewServer(logger Logger, store Store) *Server {
    return &Server{logger: logger, store: store}
}
```

### Options Pattern for Many Parameters
```go
// SMELL - Too many parameters
func NewServer(host string, port int, timeout time.Duration,
    maxConns int, logger Logger, tls bool) *Server

// OK - Options pattern
type ServerOption func(*Server)

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) { s.timeout = d }
}

func NewServer(host string, port int, opts ...ServerOption) *Server {
    s := &Server{host: host, port: port}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### Early Returns
```go
// SMELL - Deep nesting
func process(user *User) error {
    if user != nil {
        if user.Active {
            if user.Verified {
                // do work
            }
        }
    }
    return nil
}

// OK - Guard clauses
func process(user *User) error {
    if user == nil {
        return ErrNilUser
    }
    if !user.Active {
        return ErrInactiveUser
    }
    if !user.Verified {
        return ErrUnverifiedUser
    }
    // do work
    return nil
}
```

## Concurrency

### Share by Communicating
```go
// SMELL - Shared memory
type Counter struct {
    mu    sync.Mutex
    value int
}

// OK - Channel-based (when appropriate)
func counter(out chan<- int) {
    for i := 0; ; i++ {
        out <- i
    }
}
```

### Context Propagation
```go
// BROKEN - Ignores context
func fetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)
    // ...
}

// OK - Respects context
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

## Testing

### Table-Driven Tests
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

### Test Helpers
```go
func TestComplex(t *testing.T) {
    db := setupTestDB(t) // t.Cleanup registered inside
    // use db
}

func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open db: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

## Go Proverbs to Remember

1. Don't communicate by sharing memory; share memory by communicating.
2. Concurrency is not parallelism.
3. Channels orchestrate; mutexes serialize.
4. The bigger the interface, the weaker the abstraction.
5. Make the zero value useful.
6. interface{} says nothing.
7. Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.
8. A little copying is better than a little dependency.
9. Syscall must always be guarded with build tags.
10. Cgo must always be guarded with build tags.
11. Cgo is not Go.
12. With the unsafe package there are no guarantees.
13. Clear is better than clever.
14. Reflection is never clear.
15. Errors are values.
16. Don't just check errors, handle them gracefully.
17. Design the architecture, name the components, document the details.
18. Documentation is for users.
