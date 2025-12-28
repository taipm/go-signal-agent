# Go Concurrency Patterns and Rules

## Race Conditions

### The Golden Rule
**If two goroutines access the same variable, and at least one is a write, you MUST synchronize.**

### Vulnerable Patterns

```go
// 🔴 BROKEN - Data race
var counter int

func increment() {
    counter++ // NOT atomic
}

// Multiple goroutines calling increment() = RACE
```

### Safe Patterns

```go
// OK - Mutex protection
var (
    mu      sync.Mutex
    counter int
)

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// OK - Atomic operations
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

// OK - Channel-based
func counter(increment <-chan struct{}, value chan<- int) {
    var count int
    for range increment {
        count++
    }
    value <- count
}
```

### Map Race Condition

```go
// 🔴 BROKEN - Concurrent map access
var cache = make(map[string]string)

func get(key string) string {
    return cache[key] // Race!
}

func set(key, value string) {
    cache[key] = value // Race!
}

// OK - sync.Map for simple cases
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

// OK - RWMutex for complex cases
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

## Deadlocks

### Common Patterns

```go
// 🔴 BROKEN - Self-deadlock
var mu sync.Mutex

func foo() {
    mu.Lock()
    bar() // Calls Lock again!
    mu.Unlock()
}

func bar() {
    mu.Lock() // DEADLOCK
    // ...
    mu.Unlock()
}

// 🔴 BROKEN - Lock ordering violation
var mu1, mu2 sync.Mutex

func goroutine1() {
    mu1.Lock()
    mu2.Lock() // Waits for goroutine2
    // ...
}

func goroutine2() {
    mu2.Lock()
    mu1.Lock() // Waits for goroutine1 -> DEADLOCK
    // ...
}
```

### Safe Patterns

```go
// OK - Consistent lock ordering
func goroutine1() {
    mu1.Lock()
    mu2.Lock()
    // ...
    mu2.Unlock()
    mu1.Unlock()
}

func goroutine2() {
    mu1.Lock() // Same order as goroutine1
    mu2.Lock()
    // ...
    mu2.Unlock()
    mu1.Unlock()
}

// OK - Use defer for cleanup
func process() {
    mu.Lock()
    defer mu.Unlock()
    // ... all paths release lock
}
```

### Channel Deadlocks

```go
// 🔴 BROKEN - Unbuffered channel, no receiver
ch := make(chan int)
ch <- 1 // Blocks forever, no receiver

// 🔴 BROKEN - Receiving from nil channel
var ch chan int
<-ch // Blocks forever

// OK - Buffered channel
ch := make(chan int, 1)
ch <- 1 // Doesn't block

// OK - Ensure receiver exists
ch := make(chan int)
go func() {
    val := <-ch
    fmt.Println(val)
}()
ch <- 1
```

---

## Goroutine Leaks

### Common Patterns

```go
// 🔴 BROKEN - Goroutine never exits
func startWorker() {
    go func() {
        for {
            // No exit condition!
            process()
        }
    }()
}

// 🔴 BROKEN - Blocked on channel forever
func fetch(url string) <-chan Result {
    ch := make(chan Result)
    go func() {
        result, err := http.Get(url)
        if err != nil {
            return // Goroutine exits but channel never closes
        }
        ch <- result // If no receiver, goroutine leaks
    }()
    return ch
}
```

### Safe Patterns

```go
// OK - Context for cancellation
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // Clean exit
            default:
                process()
            }
        }
    }()
}

// OK - Buffered channel prevents block
func fetch(url string) <-chan Result {
    ch := make(chan Result, 1) // Buffered!
    go func() {
        defer close(ch)
        result, err := http.Get(url)
        if err != nil {
            return
        }
        ch <- result
    }()
    return ch
}

// OK - Done channel pattern
func worker(done <-chan struct{}, work <-chan Work) {
    for {
        select {
        case <-done:
            return
        case w := <-work:
            process(w)
        }
    }
}
```

---

## Channel Misuse

### Nil Channel Issues

```go
// 🔴 BROKEN - Send to nil channel blocks forever
var ch chan int
ch <- 1 // Blocks forever

// 🔴 BROKEN - Receive from nil channel blocks forever
var ch chan int
<-ch // Blocks forever

// OK - Nil channel in select disables case
var ch1, ch2 chan int
ch1 = make(chan int)

select {
case v := <-ch1:
    fmt.Println(v)
case v := <-ch2: // This case is disabled (ch2 is nil)
    fmt.Println(v)
}
```

### Closing Channels

```go
// 🔴 BROKEN - Close nil channel
var ch chan int
close(ch) // Panic!

// 🔴 BROKEN - Close already closed channel
ch := make(chan int)
close(ch)
close(ch) // Panic!

// 🔴 BROKEN - Send on closed channel
ch := make(chan int)
close(ch)
ch <- 1 // Panic!

// OK - Only sender closes, once
func producer(ch chan<- int) {
    defer close(ch) // Close when done
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

func consumer(ch <-chan int) {
    for v := range ch { // Exits when closed
        fmt.Println(v)
    }
}
```

---

## Context Propagation

### The Rules

1. Context should be the first parameter
2. Never store context in a struct
3. Pass context through the call chain
4. Check context.Done() in long operations

### Vulnerable Patterns

```go
// 🟡 SMELL - Ignores context
func fetchData(ctx context.Context, url string) ([]byte, error) {
    resp, err := http.Get(url) // Doesn't use ctx!
    // ...
}

// 🟡 SMELL - Context stored in struct
type Service struct {
    ctx context.Context // Don't do this
}
```

### Safe Patterns

```go
// OK - Use context properly
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := http.DefaultClient.Do(req)
    // ...
}

// OK - Check context in loops
func process(ctx context.Context, items []Item) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := processItem(item); err != nil {
                return err
            }
        }
    }
    return nil
}
```

---

## WaitGroup Patterns

### Vulnerable Patterns

```go
// 🔴 BROKEN - Add inside goroutine
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    go func() {
        wg.Add(1) // Race condition!
        defer wg.Done()
        work()
    }()
}
wg.Wait() // Might return before goroutines start

// 🔴 BROKEN - Forget Done
var wg sync.WaitGroup
wg.Add(1)
go func() {
    work()
    // Forgot wg.Done()!
}()
wg.Wait() // Blocks forever
```

### Safe Patterns

```go
// OK - Add before goroutine
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        work()
    }()
}
wg.Wait()

// OK - errgroup for error handling
g, ctx := errgroup.WithContext(ctx)
for _, item := range items {
    item := item // Capture
    g.Go(func() error {
        return process(ctx, item)
    })
}
if err := g.Wait(); err != nil {
    return err
}
```

---

## Detection Commands

```bash
# Run with race detector
go run -race .
go test -race ./...

# Common race patterns in code
grep -rn "go func" --include="*.go" | grep -v "test"
grep -rn "sync.Mutex" --include="*.go"
grep -rn "sync.RWMutex" --include="*.go"
```

---

## Quick Reference

| Issue | Severity | Detection |
|-------|----------|-----------|
| Data race | 🔴 BROKEN | `go test -race` |
| Deadlock | 🔴 BROKEN | Code review |
| Goroutine leak | 🔴 BROKEN | Memory profiling |
| Channel misuse | 🟡 SMELL | Code review |
| Missing context | 🟡 SMELL | Code review |
| WaitGroup misuse | 🔴 BROKEN | Code review |

## Key Takeaways

1. **Always use -race** during development and testing
2. **Context propagation** is not optional
3. **Lock ordering** must be consistent
4. **Close channels** from sender side only
5. **Add to WaitGroup** before starting goroutine
6. **Use defer** for unlocking mutexes
7. **Prefer channels** for communication
8. **Use mutexes** for shared state protection
9. **Check ctx.Done()** in long operations
10. **Profile for goroutine leaks** regularly
