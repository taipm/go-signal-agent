# Anti-Patterns - GLOBAL Knowledge

> Code smells and mistakes to avoid in ANY Go project.
> Severity: 🔴 BROKEN (bugs/crashes) | 🟡 SMELL (maintainability) | 🟠 PERF (performance)

## 🔴 Concurrency Issues

### Goroutine Leak

```go
// ❌ Bad: Goroutine never exits
func startWorker() {
    go func() {
        for {
            doWork() // No exit condition!
        }
    }()
}

// ✅ Good: Context for cancellation
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                doWork()
            }
        }
    }()
}
```

### Data Race on Map

```go
// ❌ Bad: Concurrent map access
var cache = make(map[string]string)

func Get(key string) string { return cache[key] }    // Race!
func Set(key, val string)   { cache[key] = val }     // Race!

// ✅ Good: Use sync.Map or mutex
var cache sync.Map

func Get(key string) (string, bool) {
    v, ok := cache.Load(key)
    if !ok {
        return "", false
    }
    return v.(string), true
}
```

### Send on Closed Channel

```go
// ❌ Bad: Panic if channel closed
func producer(ch chan<- int) {
    for i := 0; i < 10; i++ {
        ch <- i // Panic if consumer closes channel!
    }
}

// ✅ Good: Producer owns channel lifecycle
func producer() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch) // Producer closes
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    return ch
}
```

## 🟡 Code Smells

### Ignoring Errors

```go
// ❌ Bad: Silent failure
data, _ := json.Marshal(obj)
file.Write(data)

// ✅ Good: Handle or propagate
data, err := json.Marshal(obj)
if err != nil {
    return fmt.Errorf("marshal failed: %w", err)
}
if _, err := file.Write(data); err != nil {
    return fmt.Errorf("write failed: %w", err)
}
```

### Empty Interface Abuse

```go
// ❌ Bad: Type safety lost
func Process(data interface{}) {
    // Runtime type assertions, error-prone
}

// ✅ Good: Use generics or specific interface
func Process[T Processable](data T) {
    // Type-safe at compile time
}
```

### Magic Numbers

```go
// ❌ Bad: What is 86400?
cache.SetTTL(86400)

// ✅ Good: Named constant
const DayInSeconds = 24 * 60 * 60
cache.SetTTL(DayInSeconds)

// ✅ Better: Use time.Duration
cache.SetTTL(24 * time.Hour)
```

### Long Function

```go
// ❌ Bad: 100+ lines, multiple responsibilities
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Validate input (20 lines)
    // Query database (30 lines)
    // Transform data (25 lines)
    // Format response (25 lines)
}

// ✅ Good: Single responsibility
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    input, err := validateInput(r)
    if err != nil { ... }

    data, err := queryData(input)
    if err != nil { ... }

    result := transformData(data)
    writeResponse(w, result)
}
```

## 🟠 Performance Issues

### String Concatenation in Loop

```go
// ❌ Bad: O(n²) allocations
var s string
for _, part := range parts {
    s += part
}

// ✅ Good: O(n) with Builder
var b strings.Builder
for _, part := range parts {
    b.WriteString(part)
}
```

### Defer in Tight Loop

```go
// ❌ Bad: Defer overhead per iteration
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close() // Deferred until function returns!
    process(f)
}

// ✅ Good: Explicit close or helper function
for _, file := range files {
    if err := processFile(file); err != nil { ... }
}

func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close() // OK: closes when this function returns
    return process(f)
}
```

### Allocating in Hot Path

```go
// ❌ Bad: Allocation per request
func Handle(r *Request) {
    buf := make([]byte, 4096) // Allocates every time
    // ...
}

// ✅ Good: Reuse with sync.Pool
var bufPool = sync.Pool{
    New: func() interface{} { return make([]byte, 4096) },
}

func Handle(r *Request) {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf)
    // ...
}
```

## 🟡 Design Issues

### Package Cycle

```go
// ❌ Bad: Circular dependency
// package a imports package b
// package b imports package a

// ✅ Good: Interface in shared package
// package types (shared)
type Processor interface { Process() }

// package a
func NewA(p types.Processor) *A { ... }

// package b
func (b *B) Process() { ... } // implements types.Processor
```

### God Object

```go
// ❌ Bad: One struct does everything
type App struct {
    db       *sql.DB
    cache    *redis.Client
    logger   *log.Logger
    mailer   *smtp.Client
    // ... 20 more fields
}

func (a *App) HandleUser() { ... }
func (a *App) SendEmail() { ... }
func (a *App) CacheResult() { ... }
// ... 50 more methods

// ✅ Good: Separate concerns
type UserService struct { db *sql.DB }
type EmailService struct { mailer *smtp.Client }
type CacheService struct { cache *redis.Client }
```

---

*Last updated: 2025-12-29*
*Add new anti-patterns as you discover them*
