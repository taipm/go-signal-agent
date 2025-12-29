# Refactoring Patterns - GLOBAL Knowledge

> Successful refactoring patterns discovered across ALL projects.
> Add new patterns here when they apply universally to Go code.

## Pattern: Generic HTTP Streaming Helper

**Context**: Multiple methods with identical HTTP streaming logic but different response types

**Before**:

```go
// Duplicated in GenerateStream and ChatStream
func (c *Client) GenerateStream(...) error {
    // 60+ lines of duplicated logic:
    // - Create HTTP request
    // - Execute request
    // - Check status, read error body
    // - Scan response lines
    // - Unmarshal JSON chunks
    // - Call callback
    // - Handle done flag
}
```

**After**:

```go
// Generic helper function (not method - Go limitation)
func doStreamRequest[T any](
    c *Client,
    ctx context.Context,
    endpoint string,
    body []byte,
    callback StreamCallback,
    extract func(*T) (string, bool),
) error

// Simplified caller
func (c *Client) GenerateStream(...) error {
    req := GenerateRequest{...}
    body, _ := json.Marshal(req)
    return doStreamRequest(c, ctx, "/api/generate", body, callback,
        func(chunk *GenerateResponse) (string, bool) {
            return chunk.Response, chunk.Done
        })
}
```

**Benefits**:

- Reduced duplication from 120+ lines to 50 lines
- Single point of maintenance for streaming logic
- Type-safe with generics

**Go Insight**: Methods cannot have type parameters, use package-level functions instead.

---

## Pattern: Error Body Reader Extraction

**Context**: Same error reading pattern repeated in multiple places

**Before**:

```go
bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBodySize))
return fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(bodyBytes))
```

**After**:

```go
func readErrorBody(resp *http.Response) error {
    bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBodySize))
    return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
}
```

**Benefits**: Single source of truth, easier to modify error format, consistent error messages.

---

## Pattern: Functional Options

**Context**: Structs with many optional configuration fields

**Before**:

```go
type Client struct {
    Timeout    time.Duration
    MaxRetries int
    BaseURL    string
    Logger     Logger
}

// Caller must know all fields
client := &Client{
    Timeout:    30 * time.Second,
    MaxRetries: 3,
    BaseURL:    "https://api.example.com",
    Logger:     nil, // What's the default?
}
```

**After**:

```go
type Option func(*Client)

func WithTimeout(d time.Duration) Option {
    return func(c *Client) { c.timeout = d }
}

func WithRetries(n int) Option {
    return func(c *Client) { c.maxRetries = n }
}

func NewClient(baseURL string, opts ...Option) *Client {
    c := &Client{
        baseURL:    baseURL,
        timeout:    30 * time.Second, // sensible default
        maxRetries: 3,                // sensible default
    }
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// Clean usage
client := NewClient("https://api.example.com",
    WithTimeout(60 * time.Second),
    WithRetries(5),
)
```

**Benefits**: Self-documenting, sensible defaults, extensible without breaking changes.

---

## Pattern: Table-Driven Tests

**Context**: Multiple test cases with similar structure

**Before**:

```go
func TestAdd(t *testing.T) {
    if Add(1, 2) != 3 {
        t.Error("1+2 should be 3")
    }
    if Add(0, 0) != 0 {
        t.Error("0+0 should be 0")
    }
    if Add(-1, 1) != 0 {
        t.Error("-1+1 should be 0")
    }
}
```

**After**:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"zeros", 0, 0, 0},
        {"negative", -1, 1, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

**Benefits**: Easy to add cases, clear test names, parallel-friendly with t.Parallel().

---

## Pattern: Interface for Testability

**Context**: Code that's hard to test due to external dependencies

**Before**:

```go
func SendNotification(userID string) error {
    // Directly calls external service - hard to test
    resp, err := http.Post("https://api.slack.com/...", ...)
    // ...
}
```

**After**:

```go
type Notifier interface {
    Send(userID, message string) error
}

type SlackNotifier struct {
    client *http.Client
}

func (s *SlackNotifier) Send(userID, message string) error {
    // Real implementation
}

// In tests
type MockNotifier struct {
    SendFunc func(userID, message string) error
}

func (m *MockNotifier) Send(userID, message string) error {
    return m.SendFunc(userID, message)
}
```

**Benefits**: Testable code, swappable implementations, clear dependencies.

---

## Risk-Based Pattern Classification

### 🟢 LOW Risk Patterns (Auto-batch safe)

These patterns are mechanical transformations with no behavior change:

| Pattern | Before | After | Risk Level |
|---------|--------|-------|------------|
| Deprecated API | `ioutil.ReadAll(r)` | `io.ReadAll(r)` | 🟢 LOW |
| String Builder | `s += x` in loop | `strings.Builder` | 🟢 LOW |
| Slice Prealloc | `var s []T` | `make([]T, 0, n)` | 🟢 LOW |
| Import Cleanup | Unused imports | Remove | 🟢 LOW |
| Const Extract | Magic number `100` | `const x = 100` | 🟢 LOW |
| Formatting | Inconsistent | `gofmt` | 🟢 LOW |

### 🟡 MEDIUM Risk Patterns (Group-confirm)

These patterns change code structure but preserve logic:

| Pattern | Trigger | Risk Level |
|---------|---------|------------|
| Extract Function | Duplicate code blocks | 🟡 MEDIUM |
| Early Return | Deep nesting `if/else` | 🟡 MEDIUM |
| Error Wrapping | Raw error return | 🟡 MEDIUM |
| Interface Extract | Concrete dependency | 🟡 MEDIUM |
| Table-Driven Test | Repetitive test cases | 🟡 MEDIUM |

### 🔴 HIGH Risk Patterns (Individual-confirm)

These patterns may change behavior or API:

| Pattern | Impact | Risk Level |
|---------|--------|------------|
| Add Context param | API breaking change | 🔴 HIGH |
| Change error handling | Behavior change | 🔴 HIGH |
| Concurrency intro | Race condition risk | 🔴 HIGH |
| Algorithm change | Logic change | 🔴 HIGH |
| Security-related | Security impact | 🔴 HIGH |

---

*Last updated: 2025-12-29*
*Add new patterns as you discover them across projects*
