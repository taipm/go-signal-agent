# Interactive CLI Patterns

> "Talk is cheap. Show me the code." — Linus Torvalds

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                      Main Goroutine                           │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  for {                                                   │ │
│  │      select {                                            │ │
│  │      case <-ctx.Done():     // Shutdown signal           │ │
│  │      case input := <-inputChan:  // User input           │ │
│  │      case result := <-resultChan: // Processing result   │ │
│  │      }                                                   │ │
│  │  }                                                       │ │
│  └─────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
              ▲                              ▲
              │                              │
         ┌────┴────┐                    ┌────┴────┐
         │ Input   │                    │ Worker  │
         │ Reader  │                    │ Pool    │
         │ (1 ONLY)│                    │         │
         └─────────┘                    └─────────┘
              │                              │
              ▼                              ▼
         os.Stdin                      Engine/LLM/etc
```

**Key Principle:** Input reader là **1 goroutine duy nhất** tồn tại suốt application lifetime.

---

## Complete Template - Copy This

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func run() error {
    // 1. Setup context with signal handling
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // 2. Start input reader goroutine (SINGLE instance)
    inputChan, inputDone := startInputReader(ctx)

    // 3. Main REPL loop
    fmt.Print("> ")
    for {
        select {
        case <-ctx.Done():
            fmt.Println("\nGoodbye!")
            // Wait for input goroutine cleanup
            select {
            case <-inputDone:
            case <-time.After(100 * time.Millisecond):
            }
            return nil

        case input, ok := <-inputChan:
            if !ok {
                return nil // EOF
            }

            input = strings.TrimSpace(input)
            if input == "" {
                fmt.Print("> ")
                continue
            }

            if input == "exit" || input == "quit" {
                cancel()
                continue
            }

            // Process input
            result := process(ctx, input)
            fmt.Printf("%s\n> ", result)
        }
    }
}

// startInputReader creates a single persistent input reading goroutine
func startInputReader(ctx context.Context) (<-chan string, <-chan struct{}) {
    inputChan := make(chan string)
    done := make(chan struct{})

    go func() {
        defer close(done)
        defer close(inputChan)

        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            select {
            case inputChan <- scanner.Text():
            case <-ctx.Done():
                return
            }
        }
    }()

    return inputChan, done
}

func process(ctx context.Context, input string) string {
    // Your processing logic here
    // ALWAYS accept context for cancellation
    return strings.ToUpper(input)
}
```

---

## ❌ THẢM HỌA: Anti-Patterns

### Pattern 1: Goroutine per iteration

```go
// 🔴 BROKEN — GOROUTINE LEAK, MEMORY EXPLOSION
for {
    fmt.Print("> ")

    inputChan := make(chan string, 1)  // New channel every iteration!
    go func() {
        scanner.Scan()                  // New goroutine every iteration!
        inputChan <- scanner.Text()     // This goroutine NEVER exits
    }()

    select {
    case input := <-inputChan:
        process(input)
    case <-shutdown:
        return  // Old goroutines still alive!
    }
}
```

**Tại sao chết:**
- Iteration 1: 1 goroutine blocked on Scan()
- Iteration 2: 2 goroutines blocked
- Iteration 100: 100 goroutines blocked
- Ctrl+C: Tất cả 100 goroutines vẫn sống

### Pattern 2: Blocking in main goroutine

```go
// 🔴 BROKEN — CTRL+C KHÔNG HOẠT ĐỘNG
for {
    fmt.Print("> ")
    input, _ := reader.ReadString('\n')  // BLOCKS HERE

    select {
    case <-ctx.Done():  // NEVER REACHED while blocked above
        return
    default:
    }

    process(input)
}
```

**Tại sao chết:**
- `ReadString()` là blocking syscall
- Khi block, `select` không được evaluate
- `ctx.Done()` không được check
- Ctrl+C không có effect

### Pattern 3: Buffered channel hiding the leak

```go
// 🔴 BROKEN — Leak ẩn bởi buffer
inputChan := make(chan string, 100)  // Buffer hides the problem temporarily
for i := 0; i < 100; i++ {
    go func() {
        scanner.Scan()
        inputChan <- scanner.Text()
    }()
}
// After 100 iterations, channel full, deadlock!
```

---

## ✅ ĐÚNG: Variations

### Variation 1: With timeout per input

```go
func run(ctx context.Context) error {
    inputChan, inputDone := startInputReader(ctx)

    for {
        fmt.Print("> ")

        select {
        case <-ctx.Done():
            <-inputDone
            return nil

        case input, ok := <-inputChan:
            if !ok {
                return nil
            }
            process(input)

        case <-time.After(5 * time.Minute):
            fmt.Println("\nSession timeout")
            return nil
        }
    }
}
```

### Variation 2: With async processing

```go
func run(ctx context.Context) error {
    inputChan, inputDone := startInputReader(ctx)
    resultChan := make(chan string, 1)

    for {
        fmt.Print("> ")

        select {
        case <-ctx.Done():
            <-inputDone
            return nil

        case input, ok := <-inputChan:
            if !ok {
                return nil
            }
            // Async processing
            go func(in string) {
                result := slowProcess(ctx, in)
                select {
                case resultChan <- result:
                case <-ctx.Done():
                }
            }(input)

        case result := <-resultChan:
            fmt.Printf("%s\n", result)
        }
    }
}
```

### Variation 3: With command history

```go
type CLI struct {
    history []string
    ctx     context.Context
}

func (c *CLI) run() error {
    inputChan, inputDone := startInputReader(c.ctx)

    for {
        fmt.Printf("[%d]> ", len(c.history))

        select {
        case <-c.ctx.Done():
            <-inputDone
            return nil

        case input, ok := <-inputChan:
            if !ok {
                return nil
            }

            c.history = append(c.history, input)

            if input == "history" {
                for i, cmd := range c.history {
                    fmt.Printf("%d: %s\n", i, cmd)
                }
                continue
            }

            result := c.process(input)
            fmt.Println(result)
        }
    }
}
```

---

## Integration with Signal Engine

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Setup engine
    engine := signal.NewEngine(signal.DefaultEngineConfig())
    engine.Start()
    defer engine.Stop()

    // Register agents
    engine.GetRouter().Register("processor", NewProcessorAgent())

    // Input reader
    inputChan, inputDone := startInputReader(ctx)

    // Result channel
    resultChan := make(chan *signal.Signal, 10)

    // Main loop
    for {
        fmt.Print("> ")

        select {
        case <-ctx.Done():
            fmt.Println("\nShutting down...")
            <-inputDone
            return

        case input, ok := <-inputChan:
            if !ok {
                return
            }

            // Create and submit signal
            sig := signal.NewSignal("user_input", input).
                WithDestination("processor")

            if err := engine.Submit(sig); err != nil {
                fmt.Printf("Error: %v\n", err)
                continue
            }

        case result := <-resultChan:
            fmt.Printf("Result: %v\n", result.Payload)
        }
    }
}
```

---

## Checklist

- [ ] Input reader là **1 goroutine duy nhất** (KHÔNG tạo mới mỗi iteration)
- [ ] Input goroutine có exit path qua `ctx.Done()`
- [ ] `defer close(done)` và `defer close(inputChan)` trong input goroutine
- [ ] Main loop dùng `select` với `ctx.Done()` case
- [ ] Wait for `inputDone` trước khi exit main
- [ ] `strings.TrimSpace()` cho input
- [ ] Handle EOF (channel closed)
- [ ] Handle "exit"/"quit" commands

---

## Common Mistakes

| Mistake | Symptom | Fix |
|---------|---------|-----|
| Goroutine per iteration | Memory grows, many goroutines | Single persistent goroutine |
| Blocking stdin in main | Ctrl+C doesn't work | Move stdin to goroutine |
| Missing inputDone wait | Goroutine orphaned | Wait before exit |
| Missing defer close | Channel never closes | Add defer close() |
| No ctx.Done() in reader | Reader can't exit | Add select with ctx.Done() |

---

## Testing

```go
func TestCLI_GracefulShutdown(t *testing.T) {
    // Create test context
    ctx, cancel := context.WithCancel(context.Background())

    // Mock stdin
    r, w, _ := os.Pipe()
    oldStdin := os.Stdin
    os.Stdin = r
    defer func() { os.Stdin = oldStdin }()

    // Start CLI
    done := make(chan struct{})
    go func() {
        defer close(done)
        run(ctx)
    }()

    // Send some input
    w.WriteString("hello\n")
    time.Sleep(100 * time.Millisecond)

    // Cancel (simulate Ctrl+C)
    cancel()

    // Should exit quickly
    select {
    case <-done:
        // Good!
    case <-time.After(2 * time.Second):
        t.Fatal("CLI didn't shutdown gracefully")
    }
}
```

**Talk is cheap. Show me the code.**
