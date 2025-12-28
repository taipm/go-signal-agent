---
stepNumber: 7
nextStep: './step-08-release.md'
agent: optimizer-agent
hasBreakpoint: false
---

# Step 07: Performance Optimization

## STEP GOAL

Optimizer Agent analyzes code for performance bottlenecks and applies optimizations where beneficial.

## AGENT ACTIVATION

Load persona từ `../agents/optimizer-agent.md`

Input context:
- Code files đã qua review
- Performance requirements từ spec (if any)

## EXECUTION SEQUENCE

### 1. Optimizer Agent Introduction

```
[Optimizer Agent]

Analyzing code for optimization opportunities...

Focus areas:
1. Memory allocations
2. Concurrency patterns
3. Hot path efficiency
4. Caching opportunities
```

### 2. Analysis Phase

#### Memory Analysis
```bash
# Escape analysis
go build -gcflags='-m -m' ./... 2>&1 | grep "escapes"
```

#### Benchmark (if existing)
```bash
go test -bench=. -benchmem ./...
```

### 3. Identify Opportunities

Common patterns to check:
- [ ] Slice pre-allocation
- [ ] String building with strings.Builder
- [ ] Object pooling for frequently allocated objects
- [ ] Buffered channels
- [ ] Worker pools for concurrent work

### 4. Apply Optimizations

Example optimizations:

#### Pre-allocate Slices
```go
// Before
func Process(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// After
func Process(items []Item) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}
```

#### Use sync.Pool
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func Process(data []byte) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    // use buf
    return buf.String()
}
```

### 5. Create Benchmarks

```go
// internal/service/{entity}_service_bench_test.go
package service_test

import (
    "context"
    "testing"
)

func BenchmarkService_GetByID(b *testing.B) {
    svc := setupTestService()
    ctx := context.Background()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, _ = svc.GetByID(ctx, "test-id")
    }
}
```

### 6. Measure Improvements

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare before/after (if benchstat available)
benchstat old.txt new.txt
```

### 7. Report Results

```
[Optimizer Agent]

Optimization Report:

### Applied Optimizations

1. **Slice pre-allocation in ProcessItems**
   - Before: 15 allocs/op
   - After: 1 alloc/op
   - Improvement: 93% fewer allocations

2. **Buffer pooling in FormatResponse**
   - Before: 2048 B/op
   - After: 64 B/op
   - Improvement: 97% less memory

### Benchmarks Added
- BenchmarkService_GetByID
- BenchmarkHandler_Get

### Recommendations (Not Applied)
- Worker pool for batch processing (apply if load increases)
- Response caching (apply if same requests frequent)

### Summary
- Optimizations applied: 2
- Memory reduction: ~80%
- No performance regression in tests

Ready for Release phase.
```

## OUTPUT

```yaml
outputs:
  optimizations:
    - type: "pre-allocation"
      file: "internal/service/service.go"
      improvement: "93% fewer allocations"
    - type: "buffer-pooling"
      file: "internal/handler/handler.go"
      improvement: "97% less memory"
  benchmarks_added:
    - "internal/service/service_bench_test.go"
  recommendations:
    - "Worker pool for batch processing"
    - "Response caching"
```

## SUCCESS CRITERIA

- [ ] Code analyzed for bottlenecks
- [ ] Safe optimizations applied
- [ ] Benchmarks created for critical paths
- [ ] No regression in tests
- [ ] Ready for Release phase

## NEXT STEP

Load `./step-08-release.md`
