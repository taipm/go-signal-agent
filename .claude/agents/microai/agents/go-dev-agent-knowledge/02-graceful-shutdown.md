# Graceful Shutdown Patterns

> "Talk is cheap. Show me the code." — Linus Torvalds

---

## TL;DR - Copy This

```go
func main() {
    // 1. Root context with signal handling
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // 2. Run application
    if err := run(ctx); err != nil && err != context.Canceled {
        log.Fatal(err)
    }
}

func run(ctx context.Context) error {
    // Your application logic here
    // ALWAYS check ctx.Done() in loops
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // work
        }
    }
}
```

---

## Pattern Comparison

| Pattern | Verdict | Why |
|---------|---------|-----|
| `signal.NotifyContext` | ✅ BEST | One-liner, auto-cancel |
| `signal.Notify` + channel | ✅ OK | More control |
| No signal handling | 🔴 BROKEN | Ctrl+C won't work |
| Blocking stdin in main | 🔴 BROKEN | Can't interrupt |

---

## ❌ THẢM HỌA: Patterns SAI

### Pattern 1: Blocking stdin không thể interrupt

```go
// 🔴 BROKEN — Code sẽ KHÔNG CHẾT khi Ctrl+C
for {
    input, _ := reader.ReadString('\n')  // BLOCKS FOREVER
    process(input)
}
```

**Tại sao chết:**
- `ReadString()` là blocking syscall
- Khi SIGINT đến, goroutine vẫn đang ngủ trong kernel waiting for stdin
- `context.Done()` không thể đánh thức blocking syscall
- User phải nhấn Ctrl+C nhiều lần hoặc kill -9

### Pattern 2: Goroutine per iteration

```go
// 🔴 BROKEN — GOROUTINE LEAK
for {
    ch := make(chan string, 1)
    go func() {
        scanner.Scan()  // Goroutine này KHÔNG BAO GIỜ EXIT
        ch <- scanner.Text()
    }()
    select {
    case input := <-ch:
        process(input)
    case <-time.After(5 * time.Second):
        continue  // Goroutine cũ vẫn còn sống!
    }
}
```

**Tại sao chết:**
- Mỗi iteration spawn goroutine mới
- Goroutine cũ vẫn blocked trên `scanner.Scan()`
- Sau 100 iterations = 100 zombie goroutines
- Memory leak, file descriptor leak

### Pattern 3: Missing defer for cleanup

```go
// 🔴 BROKEN — Resource leak
func main() {
    engine.Start()

    // If panic happens here, engine.Stop() never called
    doWork()

    engine.Stop()  // May never reach this line
}
```

**Tại sao chết:**
- Panic, early return, hoặc os.Exit() = engine không cleanup
- Goroutines orphaned, connections leaked

---

## ✅ ĐÚNG: Proper Signal Handling

### Option 1: signal.NotifyContext (Go 1.16+, RECOMMENDED)

```go
func main() {
    // One-liner signal handling
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    if err := run(ctx); err != nil && err != context.Canceled {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Option 2: Manual signal.Notify (More control)

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Signal handling goroutine
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        sig := <-sigChan
        fmt.Printf("\nReceived %v, shutting down...\n", sig)
        cancel()
    }()

    if err := run(ctx); err != nil && err != context.Canceled {
        log.Fatal(err)
    }
}
```

---

## ✅ ĐÚNG: Interactive CLI with Graceful Shutdown

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Single input goroutine - KHÔNG tạo mới mỗi iteration
    inputChan := make(chan string)
    inputDone := make(chan struct{})
    go func() {
        defer close(inputDone)
        defer close(inputChan)
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            select {
            case inputChan <- scanner.Text():
            case <-ctx.Done():
                return  // ✅ Exit path exists
            }
        }
    }()

    // Main loop
    fmt.Print("> ")
    for {
        select {
        case <-ctx.Done():
            fmt.Println("\nShutting down...")
            // Wait for input goroutine to finish
            select {
            case <-inputDone:
            case <-time.After(100 * time.Millisecond):
            }
            return

        case input, ok := <-inputChan:
            if !ok {
                return  // EOF
            }
            process(input)
            fmt.Print("> ")
        }
    }
}
```

---

## ✅ ĐÚNG: Engine/Server with Graceful Shutdown

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Start engine
    engine := signal.NewEngine(config)
    engine.Start()
    defer engine.Stop()  // ✅ ALWAYS defer Stop() right after Start()

    // Start HTTP server (if needed)
    server := &http.Server{Addr: ":8080", Handler: handler}
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Printf("HTTP server error: %v", err)
        }
    }()

    // Wait for shutdown signal
    <-ctx.Done()
    fmt.Println("Shutting down...")

    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("HTTP shutdown error: %v", err)
    }

    // engine.Stop() called by defer
}
```

---

## Checklist — Verify Before Commit

- [ ] `signal.NotifyContext` hoặc `signal.Notify` + channel
- [ ] Root `context.WithCancel` được pass xuống mọi component
- [ ] Mọi goroutine có exit path qua `ctx.Done()` hoặc done channel
- [ ] `defer` cho mọi cleanup (engine.Stop(), file.Close(), conn.Close())
- [ ] Stdin/blocking I/O trong goroutine riêng, không block main
- [ ] `select` trong loops với `ctx.Done()` case
- [ ] Test với `Ctrl+C` — PHẢI exit trong < 1 second

---

## Testing Graceful Shutdown

```bash
# 1. Run the application
go run main.go

# 2. Interact normally
> hello
Response: HELLO
>

# 3. Press Ctrl+C
^C
Shutting down...

# 4. App should exit immediately (< 1 second)
# If it hangs, you have a bug!
```

### Automated Test

```go
func TestGracefulShutdown(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Start your app
    errChan := make(chan error, 1)
    go func() {
        errChan <- run(ctx)
    }()

    // Give it time to start
    time.Sleep(100 * time.Millisecond)

    // Cancel (simulate Ctrl+C)
    cancel()

    // Should exit quickly
    select {
    case err := <-errChan:
        if err != nil && err != context.Canceled {
            t.Errorf("unexpected error: %v", err)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("shutdown took too long - goroutine leak?")
    }
}
```

---

## Reference Implementations

| Status | File | Notes |
|--------|------|-------|
| ✅ ĐÚNG | `examples/chatbot-yaml/main.go:260-290` | Single input goroutine, proper select |
| ❌ SAI | `examples/chatbot/main.go` | Goroutine per iteration |
| ❌ SAI | `examples/chatbot-memory/main.go` | Blocking stdin in main |
| ❌ SAI | `examples/multi-agent-orchestrator/main.go` | Blocking stdin in main |

---

## Quick Reference

```go
// Signal handling - pick one:
ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
// OR
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

// Cleanup - ALWAYS defer:
engine.Start()
defer engine.Stop()

// Loops - ALWAYS check ctx:
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case work := <-workChan:
        process(work)
    }
}

// Stdin - NEVER block main:
go func() {
    defer close(inputChan)
    for scanner.Scan() {
        select {
        case inputChan <- scanner.Text():
        case <-ctx.Done():
            return
        }
    }
}()
```

**Talk is cheap. Show me the code.**
