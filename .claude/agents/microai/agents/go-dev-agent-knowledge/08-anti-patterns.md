# Anti-Patterns — ĐỪNG LÀM THẾ NÀY

> "Talk is cheap. Show me the code." — Linus Torvalds

---

## Severity Legend

| Icon | Severity | Meaning |
|------|----------|---------|
| 🔴 | **BROKEN** | Code sẽ CHẾT trong production |
| 🟡 | **SMELL** | Code sẽ gây problems sớm muộn |
| 🟢 | **OK** | Acceptable trong một số cases |

---

## Concurrency Anti-Patterns

### 🔴 BROKEN: Goroutine Leak trong Loop

```go
// ❌ THẢM HỌA — Memory sẽ EXPLODE
for {
    go func() {
        result := blocking_operation()  // Goroutine KHÔNG BAO GIỜ EXIT
        resultChan <- result
    }()
    time.Sleep(time.Second)
}
// Sau 1 giờ = 3600 zombie goroutines
```

**Tại sao chết:**
- Mỗi iteration tạo goroutine mới
- Blocking operation không có timeout
- Không có exit path
- Memory usage tăng vô hạn

**✅ Fix:**
```go
// Single goroutine với exit path
go func() {
    for {
        select {
        case <-ctx.Done():
            return  // ✅ Exit path
        default:
            result := blocking_operation_with_timeout(ctx)
            select {
            case resultChan <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}()
```

---

### 🔴 BROKEN: Blocking Stdin Không Thể Cancel

```go
// ❌ THẢM HỌA — Ctrl+C KHÔNG hoạt động
for {
    input, _ := reader.ReadString('\n')  // BLOCKS FOREVER
    process(input)
}
```

**Tại sao chết:**
- `ReadString()` là blocking syscall
- Ctrl+C gửi SIGINT nhưng goroutine đang ngủ trong kernel
- Signal handler không thể interrupt syscall
- User phải kill -9

**✅ Fix:**
```go
// Stdin trong goroutine riêng
inputChan := make(chan string)
go func() {
    defer close(inputChan)
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        select {
        case inputChan <- scanner.Text():
        case <-ctx.Done():
            return  // ✅ Exit path
        }
    }
}()

// Main loop
for {
    select {
    case <-ctx.Done():
        return  // ✅ Can respond to Ctrl+C
    case input := <-inputChan:
        process(input)
    }
}
```

---

### 🔴 BROKEN: Data Race

```go
// ❌ THẢM HỌA — DATA CORRUPTION
var counter int

func increment() {
    counter++  // Race condition!
}

// 10 goroutines calling increment() = UNDEFINED BEHAVIOR
```

**Tại sao chết:**
- `counter++` không phải atomic operation
- Đọc-sửa-ghi có thể interleave
- Kết quả: lost updates, corrupt data
- `go test -race` sẽ fail

**✅ Fix:**
```go
// Option 1: Mutex
var (
    mu      sync.Mutex
    counter int
)

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// Option 2: Atomic
var counter atomic.Int64

func increment() {
    counter.Add(1)
}
```

---

### 🔴 BROKEN: Concurrent Map Access

```go
// ❌ THẢM HỌA — PANIC trong production
var cache = make(map[string]string)

func get(key string) string {
    return cache[key]  // Race!
}

func set(key, value string) {
    cache[key] = value  // Race!
}
```

**Tại sao chết:**
- Go maps không thread-safe
- Concurrent read/write = undefined behavior
- Có thể panic với "concurrent map read and map write"

**✅ Fix:**
```go
// Option 1: sync.Map (simple cases)
var cache sync.Map

func get(key string) (string, bool) {
    v, ok := cache.Load(key)
    if !ok {
        return "", false
    }
    return v.(string), true
}

func set(key, value string) {
    cache.Store(key, value)
}

// Option 2: RWMutex (complex cases)
var (
    mu    sync.RWMutex
    cache = make(map[string]string)
)

func get(key string) string {
    mu.RLock()
    defer mu.RUnlock()
    return cache[key]
}

func set(key, value string) {
    mu.Lock()
    defer mu.Unlock()
    cache[key] = value
}
```

---

## Resource Management Anti-Patterns

### 🔴 BROKEN: Missing defer for Cleanup

```go
// ❌ THẢM HỌA — Resource leak
func process() error {
    engine.Start()

    result, err := doWork()
    if err != nil {
        return err  // engine.Stop() KHÔNG ĐƯỢC GỌI!
    }

    engine.Stop()
    return nil
}
```

**Tại sao chết:**
- Early return = cleanup bị skip
- Panic = cleanup bị skip
- Resources leaked: goroutines, connections, file handles

**✅ Fix:**
```go
func process() error {
    engine.Start()
    defer engine.Stop()  // ✅ LUÔN được gọi

    result, err := doWork()
    if err != nil {
        return err  // defer still runs
    }

    return nil
}
```

---

### 🟡 SMELL: Ignoring Error

```go
// ❌ SMELL — Silent failure, khó debug
result, _ := riskyOperation()
file.Close()  // Error ignored
json.Unmarshal(data, &obj)  // Error ignored
```

**Tại sao có vấn đề:**
- Errors bị nuốt, không ai biết có lỗi
- Bugs rất khó trace
- Production issues không có logs

**✅ Fix:**
```go
result, err := riskyOperation()
if err != nil {
    return fmt.Errorf("risky operation: %w", err)
}

if err := file.Close(); err != nil {
    log.Printf("warning: close file: %v", err)
}

if err := json.Unmarshal(data, &obj); err != nil {
    return fmt.Errorf("unmarshal: %w", err)
}
```

---

### 🟡 SMELL: context.Background() Everywhere

```go
// ❌ SMELL — Không thể cancel từ outside
func process() {
    ctx := context.Background()  // Orphan context
    callAPI(ctx)
    queryDB(ctx)
}
```

**Tại sao có vấn đề:**
- Không thể cancel long-running operations
- Graceful shutdown không hoạt động
- Resources không được release kịp thời

**✅ Fix:**
```go
func process(ctx context.Context) error {  // Accept context
    if err := callAPI(ctx); err != nil {
        return err
    }
    return queryDB(ctx)
}

// Caller controls lifetime
ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
defer cancel()
process(ctx)
```

---

## Design Anti-Patterns

### 🟡 SMELL: Magic Numbers

```go
// ❌ SMELL — Không ai biết 100 là gì
if len(users) > 100 {
    return errors.New("too many users")
}

time.Sleep(5 * time.Second)  // Why 5?

buffer := make([]byte, 4096)  // Why 4096?
```

**✅ Fix:**
```go
const (
    MaxUsersPerRequest = 100
    RetryDelay         = 5 * time.Second
    DefaultBufferSize  = 4096
)

if len(users) > MaxUsersPerRequest {
    return fmt.Errorf("exceeded max users: %d > %d", len(users), MaxUsersPerRequest)
}

time.Sleep(RetryDelay)
buffer := make([]byte, DefaultBufferSize)
```

---

### 🟡 SMELL: Hardcoded Secrets

```go
// ❌ SMELL — Security nightmare
const apiKey = "sk-1234567890abcdef"  // In git history FOREVER
const dbPassword = "admin123"

client := NewClient(apiKey)
db := Connect("user:admin123@localhost/db")
```

**✅ Fix:**
```go
func main() {
    apiKey := os.Getenv("API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable required")
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        log.Fatal("DB_PASSWORD environment variable required")
    }

    client := NewClient(apiKey)
    db := Connect(fmt.Sprintf("user:%s@localhost/db", dbPassword))
}
```

---

### 🟡 SMELL: Empty Interface Abuse

```go
// ❌ SMELL — Type safety đi đâu rồi?
func process(data interface{}) interface{} {
    // No type information
    // Runtime panics waiting to happen
    return data
}
```

**✅ Fix:**
```go
// Option 1: Concrete types
func processUser(user *User) (*Result, error) {
    // Type-safe, IDE autocomplete, compile-time checks
    return &Result{UserID: user.ID}, nil
}

// Option 2: Generics (Go 1.18+)
func process[T Processable](data T) (T, error) {
    // Type-safe with constraints
    return data, nil
}

// Option 3: Interface with methods
type Processor interface {
    Process() error
}

func process(p Processor) error {
    return p.Process()
}
```

---

## Channel Anti-Patterns

### 🔴 BROKEN: Send on Closed Channel

```go
// ❌ THẢM HỌA — PANIC
ch := make(chan int)
close(ch)
ch <- 1  // PANIC: send on closed channel
```

**Rule:** Only sender should close channel. Never close from receiver.

**✅ Fix:**
```go
func producer(ch chan<- int) {
    defer close(ch)  // Sender closes
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

func consumer(ch <-chan int) {
    for v := range ch {  // Exits when closed
        fmt.Println(v)
    }
}
```

---

### 🔴 BROKEN: Receive from Nil Channel

```go
// ❌ THẢM HỌA — BLOCKS FOREVER
var ch chan int
<-ch  // Blocks forever, deadlock
```

**✅ Fix:**
```go
ch := make(chan int)  // Initialize!
// OR
if ch != nil {
    <-ch
}
```

---

## Quick Reference Table

| Anti-Pattern | Severity | Detection | Fix |
|--------------|----------|-----------|-----|
| Goroutine per iteration | 🔴 | runtime.NumGoroutine() grows | Single persistent goroutine |
| Blocking stdin | 🔴 | Ctrl+C doesn't work | Goroutine + select |
| Data race | 🔴 | `go test -race` | Mutex/Atomic |
| Map race | 🔴 | panic in prod | sync.Map/RWMutex |
| Missing defer | 🔴 | Resource leaks | Always defer cleanup |
| Ignored error | 🟡 | Silent failures | Handle or return |
| context.Background | 🟡 | Can't cancel | Propagate context |
| Magic numbers | 🟡 | Code review | Named constants |
| Hardcoded secrets | 🟡 | Security scan | Environment vars |
| Empty interface | 🟡 | Type assertions | Generics/Interfaces |
| Send closed | 🔴 | Panic | Only sender closes |
| Nil channel | 🔴 | Deadlock | Initialize channel |

---

**Talk is cheap. Show me the code.**
