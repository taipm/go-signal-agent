# Performance Patterns - Optimizer Agent Knowledge

**Version:** 1.0.0
**Agent:** Optimizer Agent

---

## TL;DR

- Pre-allocate slices khi biết size
- sync.Pool cho frequent allocations
- strings.Builder cho string concatenation
- Worker pools cho concurrent processing
- Benchmark trước và sau optimize

---

## 1. Memory Optimization

### Slice Pre-allocation

```go
// ❌ WRONG - Multiple allocations
func Process(n int) []int {
    var result []int
    for i := 0; i < n; i++ {
        result = append(result, i)  // may reallocate 10+ times for n=1000
    }
    return result
}

// ✅ CORRECT - Single allocation
func Process(n int) []int {
    result := make([]int, 0, n)
    for i := 0; i < n; i++ {
        result = append(result, i)
    }
    return result
}

// ✅ ALTERNATIVE - Direct assignment
func Process(n int) []int {
    result := make([]int, n)
    for i := 0; i < n; i++ {
        result[i] = i
    }
    return result
}
```

### sync.Pool for Reusable Objects

```go
// Buffer pool for reducing allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    // Use buffer
    buf.Write(data)
    return buf.String()
}
```

### Avoid Unnecessary Copying

```go
// ❌ WRONG - Unnecessary copy
func Process(data []byte) {
    dataCopy := make([]byte, len(data))
    copy(dataCopy, data)
    // use dataCopy
}

// ✅ CORRECT - Use original if not modifying
func Process(data []byte) {
    // Use data directly if read-only
}

// ✅ CORRECT - Use slice if only need portion
func ProcessFirst100(data []byte) {
    portion := data[:min(100, len(data))]  // no allocation
    // use portion
}
```

---

## 2. String Optimization

### String Builder

```go
// ❌ WRONG - O(n²) complexity
func BuildString(items []string) string {
    result := ""
    for _, item := range items {
        result += item  // allocates new string each time
    }
    return result
}

// ✅ CORRECT - O(n) complexity
func BuildString(items []string) string {
    var b strings.Builder
    b.Grow(estimateSize(items))  // optional: pre-grow

    for _, item := range items {
        b.WriteString(item)
    }
    return b.String()
}

// ✅ ALTERNATIVE - strings.Join
func BuildString(items []string) string {
    return strings.Join(items, "")
}
```

### Avoid []byte to string Conversion

```go
// ❌ WRONG - Unnecessary allocations
func Process(data []byte) bool {
    str := string(data)  // allocation
    return strings.Contains(str, "error")
}

// ✅ CORRECT - Work with bytes directly
func Process(data []byte) bool {
    return bytes.Contains(data, []byte("error"))
}
```

---

## 3. Concurrency Patterns

### Worker Pool

```go
type Task struct {
    ID   int
    Data interface{}
}

type Result struct {
    TaskID int
    Output interface{}
    Err    error
}

func WorkerPool(ctx context.Context, tasks <-chan Task, numWorkers int) <-chan Result {
    results := make(chan Result, numWorkers)

    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case task, ok := <-tasks:
                    if !ok {
                        return
                    }
                    output, err := processTask(ctx, task)
                    results <- Result{TaskID: task.ID, Output: output, Err: err}
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}
```

### Fan-Out/Fan-In

```go
func FanOutFanIn(ctx context.Context, inputs []int, workers int) []int {
    inputCh := make(chan int, len(inputs))
    resultCh := make(chan int, len(inputs))

    // Fan out - spawn workers
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for input := range inputCh {
                select {
                case resultCh <- process(input):
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    // Send inputs
    go func() {
        for _, input := range inputs {
            inputCh <- input
        }
        close(inputCh)
    }()

    // Fan in - collect results
    go func() {
        wg.Wait()
        close(resultCh)
    }()

    var results []int
    for result := range resultCh {
        results = append(results, result)
    }

    return results
}
```

### Semaphore Pattern

```go
func ProcessWithLimit(ctx context.Context, items []Item, maxConcurrent int) error {
    sem := make(chan struct{}, maxConcurrent)
    errCh := make(chan error, len(items))

    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()

            // Acquire semaphore
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                errCh <- ctx.Err()
                return
            }

            if err := processItem(ctx, item); err != nil {
                errCh <- err
            }
        }(item)
    }

    wg.Wait()
    close(errCh)

    // Check for errors
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    return nil
}
```

---

## 4. Database Optimization

### Connection Pool

```go
func SetupDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // Pool configuration
    db.SetMaxOpenConns(25)                  // Max open connections
    db.SetMaxIdleConns(5)                   // Max idle connections
    db.SetConnMaxLifetime(5 * time.Minute)  // Max connection lifetime
    db.SetConnMaxIdleTime(1 * time.Minute)  // Max idle time

    return db, nil
}
```

### Batch Operations

```go
// ❌ WRONG - N queries
func InsertUsers(ctx context.Context, db *sql.DB, users []User) error {
    for _, user := range users {
        _, err := db.ExecContext(ctx,
            "INSERT INTO users (id, name) VALUES ($1, $2)",
            user.ID, user.Name,
        )
        if err != nil {
            return err
        }
    }
    return nil
}

// ✅ CORRECT - Batch insert
func InsertUsers(ctx context.Context, db *sql.DB, users []User) error {
    if len(users) == 0 {
        return nil
    }

    // Build batch query
    valueStrings := make([]string, 0, len(users))
    valueArgs := make([]interface{}, 0, len(users)*2)

    for i, user := range users {
        valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
        valueArgs = append(valueArgs, user.ID, user.Name)
    }

    query := fmt.Sprintf(
        "INSERT INTO users (id, name) VALUES %s",
        strings.Join(valueStrings, ","),
    )

    _, err := db.ExecContext(ctx, query, valueArgs...)
    return err
}
```

### Prepared Statements

```go
type UserRepo struct {
    db         *sql.DB
    insertStmt *sql.Stmt
    selectStmt *sql.Stmt
}

func NewUserRepo(db *sql.DB) (*UserRepo, error) {
    insertStmt, err := db.Prepare("INSERT INTO users (id, name) VALUES ($1, $2)")
    if err != nil {
        return nil, err
    }

    selectStmt, err := db.Prepare("SELECT id, name FROM users WHERE id = $1")
    if err != nil {
        insertStmt.Close()
        return nil, err
    }

    return &UserRepo{
        db:         db,
        insertStmt: insertStmt,
        selectStmt: selectStmt,
    }, nil
}

func (r *UserRepo) Insert(ctx context.Context, user *User) error {
    _, err := r.insertStmt.ExecContext(ctx, user.ID, user.Name)
    return err
}

func (r *UserRepo) Close() {
    r.insertStmt.Close()
    r.selectStmt.Close()
}
```

---

## 5. HTTP Optimization

### Connection Reuse

```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}

func Fetch(url string) ([]byte, error) {
    resp, err := httpClient.Get(url)  // Reuses connections
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

### Response Streaming

```go
// ❌ WRONG - Load entire response into memory
func FetchLarge(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)  // Dangerous for large responses
}

// ✅ CORRECT - Stream processing
func ProcessLarge(url string, w io.Writer) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    _, err = io.Copy(w, resp.Body)  // Streams without loading all
    return err
}
```

---

## 6. Benchmarking

### Basic Benchmark

```go
func BenchmarkProcess(b *testing.B) {
    data := generateTestData(1000)

    b.ResetTimer()  // Exclude setup time
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}

func BenchmarkProcess_Parallel(b *testing.B) {
    data := generateTestData(1000)

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            Process(data)
        }
    })
}
```

### Memory Benchmark

```go
func BenchmarkAllocation(b *testing.B) {
    b.ReportAllocs()  // Report allocations

    for i := 0; i < b.N; i++ {
        result := make([]int, 1000)
        _ = result
    }
}
```

### Comparison Benchmark

```go
func BenchmarkStringConcat(b *testing.B) {
    items := []string{"a", "b", "c", "d", "e"}

    b.Run("Plus", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := ""
            for _, s := range items {
                result += s
            }
        }
    })

    b.Run("Builder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            for _, s := range items {
                builder.WriteString(s)
            }
            _ = builder.String()
        }
    })

    b.Run("Join", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = strings.Join(items, "")
        }
    })
}
```

---

## 7. Commands

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Profile CPU
go test -bench=BenchmarkProcess -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Profile Memory
go test -bench=BenchmarkProcess -memprofile=mem.prof
go tool pprof mem.prof

# Escape analysis
go build -gcflags='-m -m' ./...

# Trace
go test -trace=trace.out
go tool trace trace.out
```

---

## Quick Reference

| Problem | Solution |
|---------|----------|
| Slice growing | Pre-allocate with `make([]T, 0, n)` |
| Frequent allocations | `sync.Pool` |
| String concatenation | `strings.Builder` |
| Concurrent processing | Worker pools |
| DB queries | Batch operations, prepared statements |
| HTTP connections | Reuse client |

---

## Related Knowledge

- [02-profiling.md](./02-profiling.md) - Profiling tools
- [../shared/01-go-fundamentals.md](../shared/01-go-fundamentals.md) - Go basics
