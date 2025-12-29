# Go Idioms - GLOBAL Knowledge

> Accumulated Go best practices from Effective Go and real-world experience.
> This file is shared across ALL projects using go-refactor agent.

## Error Handling

### Early Return Pattern

```go
// ❌ Bad: Deep nesting
func process(data []byte) error {
    if data != nil {
        if len(data) > 0 {
            if isValid(data) {
                // actual logic
            }
        }
    }
    return nil
}

// ✅ Good: Early returns
func process(data []byte) error {
    if data == nil {
        return errors.New("data is nil")
    }
    if len(data) == 0 {
        return errors.New("data is empty")
    }
    if !isValid(data) {
        return errors.New("data is invalid")
    }
    // actual logic
    return nil
}
```

### Error Wrapping

```go
// ✅ Good: Add context when wrapping
if err != nil {
    return fmt.Errorf("failed to process user %s: %w", userID, err)
}
```

## Naming Conventions

### Variable Names

- Short names for small scopes: `i`, `n`, `r`, `w`
- Descriptive names for larger scopes: `userCount`, `responseWriter`
- Acronyms: `URL`, `HTTP`, `ID` (not `Url`, `Http`, `Id`)

### Interface Names

- Single method: Add `-er` suffix: `Reader`, `Writer`, `Closer`
- Multiple methods: Descriptive noun: `ReadWriter`, `FileSystem`

### Package Names

- Lowercase, single word: `http`, `json`, `signal`
- No underscores, no mixedCaps
- Short but descriptive

## Concurrency

### Channel Patterns

```go
// ✅ Good: Always check context in loops
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case item := <-items:
        process(item)
    }
}
```

### Goroutine Cleanup

```go
// ✅ Good: Ensure goroutine can exit
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return // Clean exit
        case job, ok := <-jobs:
            if !ok {
                return // Channel closed
            }
            process(job)
        }
    }
}
```

## Memory Efficiency

### Slice Preallocation

```go
// ❌ Bad: Multiple allocations
var result []int
for i := 0; i < n; i++ {
    result = append(result, i)
}

// ✅ Good: Single allocation
result := make([]int, 0, n)
for i := 0; i < n; i++ {
    result = append(result, i)
}
```

### String Building

```go
// ❌ Bad: O(n²) string concatenation
var s string
for _, part := range parts {
    s += part
}

// ✅ Good: O(n) with strings.Builder
var b strings.Builder
for _, part := range parts {
    b.WriteString(part)
}
s := b.String()
```

## Interface Design

### Keep Interfaces Small

```go
// ✅ Good: Small, focused interface
type Reader interface {
    Read(p []byte) (n int, err error)
}

// ❌ Bad: Too large, hard to implement
type FileSystem interface {
    Open(name string) (File, error)
    Create(name string) (File, error)
    Remove(name string) error
    Rename(old, new string) error
    Stat(name string) (FileInfo, error)
    // ... 10 more methods
}
```

### Accept Interfaces, Return Structs

```go
// ✅ Good: Flexible input, concrete output
func ProcessData(r io.Reader) (*Result, error) {
    // Can accept any Reader
    return &Result{...}, nil
}
```

## Context Usage

### Always First Parameter

```go
// ✅ Good
func DoSomething(ctx context.Context, arg1 string) error

// ❌ Bad
func DoSomething(arg1 string, ctx context.Context) error
```

### Don't Store in Structs

```go
// ❌ Bad: Context in struct
type Server struct {
    ctx context.Context
}

// ✅ Good: Pass to methods
func (s *Server) Handle(ctx context.Context, req *Request) error
```

---

*Last updated: 2025-12-29*
*Source: Effective Go, Go Code Review Comments, Real-world experience*
