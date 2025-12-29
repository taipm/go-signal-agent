# Code Review Checklist - Reviewer Agent Knowledge

**Version:** 1.0.0
**Agent:** Reviewer Agent

---

## TL;DR

- Code correctness và logic
- Error handling đầy đủ
- Concurrency safety
- Resource cleanup
- Security vulnerabilities
- Style consistency

---

## 1. Code Correctness

### Logic Errors

```go
// ❌ WRONG - Off-by-one error
for i := 0; i <= len(items); i++ {  // should be < not <=
    process(items[i])
}

// ❌ WRONG - Nil pointer dereference
func GetName(user *User) string {
    return user.Name  // user could be nil
}

// ✅ CORRECT
func GetName(user *User) string {
    if user == nil {
        return ""
    }
    return user.Name
}

// ❌ WRONG - Incorrect comparison
if user.Status = "active" {  // assignment, not comparison
}

// ✅ CORRECT
if user.Status == "active" {
}
```

### Boundary Conditions

```go
// ❌ WRONG - No bounds check
func GetFirst(items []Item) Item {
    return items[0]  // panics if empty
}

// ✅ CORRECT
func GetFirst(items []Item) (Item, error) {
    if len(items) == 0 {
        return Item{}, errors.New("empty slice")
    }
    return items[0], nil
}
```

---

## 2. Error Handling

### Missing Error Checks

```go
// ❌ WRONG - Ignoring error
result, _ := doSomething()
json.Unmarshal(data, &obj)
file.Close()

// ✅ CORRECT
result, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}

if err := json.Unmarshal(data, &obj); err != nil {
    return fmt.Errorf("unmarshal: %w", err)
}

if err := file.Close(); err != nil {
    log.Printf("close file: %v", err)
}
```

### Error Context

```go
// ❌ WRONG - No context
if err != nil {
    return err
}

// ❌ WRONG - Losing error chain
if err != nil {
    return errors.New("operation failed")
}

// ✅ CORRECT - Wrap with context
if err != nil {
    return fmt.Errorf("get user %s: %w", userID, err)
}
```

---

## 3. Concurrency Safety

### Race Conditions

```go
// ❌ WRONG - Race condition on map
var cache = make(map[string]interface{})

func Get(key string) interface{} {
    return cache[key]  // concurrent read
}

func Set(key string, value interface{}) {
    cache[key] = value  // concurrent write
}

// ✅ CORRECT - sync.RWMutex
type SafeCache struct {
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (c *SafeCache) Get(key string) interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.cache[key]
}

func (c *SafeCache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[key] = value
}

// ✅ ALTERNATIVE - sync.Map
var cache sync.Map

func Get(key string) (interface{}, bool) {
    return cache.Load(key)
}

func Set(key string, value interface{}) {
    cache.Store(key, value)
}
```

### Goroutine Leaks

```go
// ❌ WRONG - Goroutine leak
func Process(ctx context.Context) {
    ch := make(chan Result)

    go func() {
        result := expensiveOperation()
        ch <- result  // blocks forever if ctx canceled
    }()

    select {
    case result := <-ch:
        return result
    case <-ctx.Done():
        return nil  // goroutine still running!
    }
}

// ✅ CORRECT - Respect context cancellation
func Process(ctx context.Context) (*Result, error) {
    ch := make(chan *Result, 1)  // buffered
    errCh := make(chan error, 1)

    go func() {
        result, err := expensiveOperation(ctx)
        if err != nil {
            errCh <- err
            return
        }
        ch <- result
    }()

    select {
    case result := <-ch:
        return result, nil
    case err := <-errCh:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Channel Safety

```go
// ❌ WRONG - Closing from multiple goroutines
func Worker(ch chan int) {
    close(ch)  // panic if already closed
}

// ✅ CORRECT - sync.Once for close
type SafeChannel struct {
    ch       chan int
    once     sync.Once
}

func (s *SafeChannel) Close() {
    s.once.Do(func() {
        close(s.ch)
    })
}
```

---

## 4. Resource Management

### File Handling

```go
// ❌ WRONG - Resource leak
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    // f.Close() never called!
    return io.ReadAll(f)
}

// ✅ CORRECT - defer close
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    return io.ReadAll(f)
}
```

### Database Connections

```go
// ❌ WRONG - Connection leak
func Query(db *sql.DB) {
    rows, _ := db.Query("SELECT * FROM users")
    // rows never closed!
    for rows.Next() {
        // process
    }
}

// ✅ CORRECT
func Query(ctx context.Context, db *sql.DB) error {
    rows, err := db.QueryContext(ctx, "SELECT * FROM users")
    if err != nil {
        return fmt.Errorf("query: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        // process
    }

    if err := rows.Err(); err != nil {
        return fmt.Errorf("iterate rows: %w", err)
    }

    return nil
}
```

### HTTP Response Body

```go
// ❌ WRONG - Body not closed
func Fetch(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    // resp.Body not closed!
    return io.ReadAll(resp.Body)
}

// ✅ CORRECT
func Fetch(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("http get: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}
```

---

## 5. Security Issues

### Hardcoded Secrets

```go
// ❌ WRONG
const apiKey = "sk_live_abc123"
const password = "admin123"
var jwtSecret = []byte("my-secret-key")

// ✅ CORRECT
func GetAPIKey() string {
    key := os.Getenv("API_KEY")
    if key == "" {
        log.Fatal("API_KEY not set")
    }
    return key
}
```

### SQL Injection

```go
// ❌ WRONG
query := "SELECT * FROM users WHERE name = '" + name + "'"
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)

// ✅ CORRECT
query := "SELECT * FROM users WHERE name = $1"
db.QueryContext(ctx, query, name)
```

### Path Traversal

```go
// ❌ WRONG
func ServeFile(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Query().Get("file")
    http.ServeFile(w, r, "/uploads/"+filename)
    // filename = "../../../etc/passwd" -> exposes system files
}

// ✅ CORRECT
func ServeFile(w http.ResponseWriter, r *http.Request) {
    filename := r.URL.Query().Get("file")

    // Clean the path
    cleaned := filepath.Clean(filename)

    // Ensure it doesn't escape uploads directory
    if strings.HasPrefix(cleaned, "..") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    fullPath := filepath.Join("/uploads", cleaned)

    // Double-check it's still under uploads
    if !strings.HasPrefix(fullPath, "/uploads/") {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    http.ServeFile(w, r, fullPath)
}
```

---

## 6. Performance Issues

### Inefficient String Concatenation

```go
// ❌ WRONG - O(n²) string concatenation
func BuildString(items []string) string {
    result := ""
    for _, item := range items {
        result += item + ","  // allocates new string each time
    }
    return result
}

// ✅ CORRECT - strings.Builder
func BuildString(items []string) string {
    var builder strings.Builder
    for i, item := range items {
        if i > 0 {
            builder.WriteByte(',')
        }
        builder.WriteString(item)
    }
    return builder.String()
}

// ✅ ALTERNATIVE - strings.Join
func BuildString(items []string) string {
    return strings.Join(items, ",")
}
```

### Unnecessary Allocations

```go
// ❌ WRONG - Slice grows repeatedly
func Process(n int) []int {
    var result []int
    for i := 0; i < n; i++ {
        result = append(result, i)  // may reallocate multiple times
    }
    return result
}

// ✅ CORRECT - Pre-allocate
func Process(n int) []int {
    result := make([]int, 0, n)  // pre-allocate capacity
    for i := 0; i < n; i++ {
        result = append(result, i)
    }
    return result
}
```

---

## 7. Style Issues

### Naming Conventions

```go
// ❌ WRONG
var user_name string      // no underscores
func GetUserID() {}       // inconsistent with getUserId elsewhere
const maxRetry = 3        // should be exported MaxRetry or unexported maxRetry

// ✅ CORRECT
var userName string
func GetUserID() {}
const maxRetry = 3        // unexported constant
const MaxRetry = 3        // exported constant
```

### Import Organization

```go
// ❌ WRONG - Unorganized
import (
    "myproject/internal/model"
    "fmt"
    "github.com/gin-gonic/gin"
    "context"
)

// ✅ CORRECT - Grouped
import (
    "context"
    "fmt"

    "github.com/gin-gonic/gin"

    "myproject/internal/model"
)
```

---

## 8. Review Commands

```bash
# Static analysis
go vet ./...
golangci-lint run

# Race detection
go test -race ./...

# Build check
go build ./...

# Test coverage
go test -cover ./...
```

---

## Quick Checklist

### Must Check

- [ ] All errors handled
- [ ] No race conditions (`go test -race`)
- [ ] Resources closed (files, connections, response bodies)
- [ ] No hardcoded secrets
- [ ] No SQL injection
- [ ] Context respected in goroutines

### Should Check

- [ ] Error messages have context
- [ ] Nil pointer checks where needed
- [ ] Slice bounds validated
- [ ] Efficient string operations
- [ ] Pre-allocated slices where possible
- [ ] Consistent naming

### Review Output Format

```markdown
## Review: [PR/File Name]

### CRITICAL (Must Fix)
1. **Race condition** in `cache.go:45`
   - Map accessed without mutex
   - Fix: Use sync.RWMutex

### HIGH (Should Fix)
1. **Error ignored** in `handler.go:123`
   - `json.Unmarshal` error not checked

### MEDIUM (Consider)
1. **Performance** in `service.go:78`
   - String concatenation in loop
   - Consider strings.Builder

### LOW (Style)
1. Import grouping inconsistent in `main.go`
```

---

## Related Knowledge

- [02-static-analysis.md](./02-static-analysis.md) - Analysis tools
- [../shared/02-error-patterns.md](../shared/02-error-patterns.md) - Error handling
- [../security/01-owasp-top10.md](../security/01-owasp-top10.md) - Security checks
