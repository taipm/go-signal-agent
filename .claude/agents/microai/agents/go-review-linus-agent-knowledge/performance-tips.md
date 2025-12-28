# Go Performance Guidelines

## Allocation Optimization

### Pre-allocate Slices

```go
// 🟡 SMELL - Grows dynamically, causes allocations
var result []int
for i := 0; i < 1000; i++ {
    result = append(result, i)
}

// OK - Pre-allocated
result := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    result = append(result, i)
}

// OK - Direct assignment if size known
result := make([]int, 1000)
for i := 0; i < 1000; i++ {
    result[i] = i
}
```

### Pre-allocate Maps

```go
// 🟡 SMELL - Grows dynamically
m := make(map[string]int)
for _, item := range items {
    m[item.Key] = item.Value
}

// OK - Pre-allocated
m := make(map[string]int, len(items))
for _, item := range items {
    m[item.Key] = item.Value
}
```

### Use sync.Pool for Frequent Allocations

```go
// 🟡 SMELL - Allocates on every call
func process() *Buffer {
    buf := new(Buffer)
    // use buffer
    return buf
}

// OK - Pool reuse
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(Buffer)
    },
}

func process() {
    buf := bufferPool.Get().(*Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    // use buffer
}
```

---

## String Operations

### Use strings.Builder for Concatenation

```go
// 🟡 SMELL - O(n²) allocations
var result string
for _, s := range strings {
    result += s
}

// OK - Single allocation
var builder strings.Builder
for _, s := range strings {
    builder.WriteString(s)
}
result := builder.String()

// OK - For known size
var builder strings.Builder
builder.Grow(totalSize) // Pre-allocate
for _, s := range strings {
    builder.WriteString(s)
}
```

### Avoid fmt.Sprintf in Hot Paths

```go
// 🟡 SMELL - Slow in loops
for i := 0; i < 1000000; i++ {
    key := fmt.Sprintf("key_%d", i)
}

// OK - strconv is faster
for i := 0; i < 1000000; i++ {
    key := "key_" + strconv.Itoa(i)
}
```

### Use []byte Instead of string When Possible

```go
// 🟡 SMELL - String conversion overhead
func hash(s string) []byte {
    h := sha256.Sum256([]byte(s))
    return h[:]
}

// OK - Accept []byte directly
func hash(data []byte) []byte {
    h := sha256.Sum256(data)
    return h[:]
}
```

---

## Interface Overhead

### Avoid interface{} in Hot Paths

```go
// 🟡 SMELL - Boxing overhead
func sum(values []interface{}) int {
    var total int
    for _, v := range values {
        total += v.(int)
    }
    return total
}

// OK - Type-specific
func sum(values []int) int {
    var total int
    for _, v := range values {
        total += v
    }
    return total
}
```

### Use Generics for Type-Safe Performance

```go
// OK - Go 1.18+ generics
func sum[T constraints.Integer](values []T) T {
    var total T
    for _, v := range values {
        total += v
    }
    return total
}
```

---

## Slice Operations

### Avoid Slice Copy in Loops

```go
// 🟡 SMELL - Creates new slice each iteration
for _, item := range items {
    filtered := append([]Item{}, filtered...)
    // ...
}

// OK - In-place filtering
n := 0
for _, item := range items {
    if keepItem(item) {
        items[n] = item
        n++
    }
}
items = items[:n]
```

### Use copy() Efficiently

```go
// 🟡 SMELL - Manual loop copy
for i, v := range src {
    dst[i] = v
}

// OK - Built-in copy
copy(dst, src)
```

### Slice Header Reuse

```go
// OK - Reuse slice (set length to 0, keep capacity)
slice = slice[:0]
for _, item := range items {
    slice = append(slice, item)
}
```

---

## Goroutine Efficiency

### Limit Goroutine Creation

```go
// 🟡 SMELL - Unbounded goroutines
for _, url := range urls {
    go fetch(url) // Could create millions
}

// OK - Worker pool pattern
sem := make(chan struct{}, 10) // Limit to 10 concurrent
for _, url := range urls {
    sem <- struct{}{}
    go func(u string) {
        defer func() { <-sem }()
        fetch(u)
    }(url)
}
```

### Use sync.WaitGroup Correctly

```go
// 🟡 SMELL - WaitGroup in goroutine
for _, item := range items {
    go func(i Item) {
        wg.Add(1) // Race condition!
        defer wg.Done()
        process(i)
    }(item)
}

// OK - Add before goroutine
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        process(i)
    }(item)
}
```

---

## Memory Layout

### Struct Field Ordering

```go
// 🟡 SMELL - Poor alignment (64 bytes due to padding)
type BadStruct struct {
    a bool   // 1 byte + 7 padding
    b int64  // 8 bytes
    c bool   // 1 byte + 7 padding
    d int64  // 8 bytes
}

// OK - Ordered by size (24 bytes)
type GoodStruct struct {
    b int64  // 8 bytes
    d int64  // 8 bytes
    a bool   // 1 byte
    c bool   // 1 byte + 6 padding
}
```

### Use Pointers Wisely

```go
// 🟡 SMELL - Pointer to small struct causes heap allocation
func newPoint() *Point { // Might escape to heap
    return &Point{X: 1, Y: 2}
}

// OK - Return by value for small structs
func newPoint() Point { // Stays on stack
    return Point{X: 1, Y: 2}
}
```

---

## I/O Operations

### Buffer I/O

```go
// 🟡 SMELL - Unbuffered writes
for _, line := range lines {
    file.WriteString(line)
}

// OK - Buffered writes
writer := bufio.NewWriter(file)
for _, line := range lines {
    writer.WriteString(line)
}
writer.Flush()
```

### Use io.Copy for Large Data

```go
// 🟡 SMELL - Load entire file into memory
data, _ := ioutil.ReadAll(src)
dst.Write(data)

// OK - Stream copy
io.Copy(dst, src)
```

---

## Benchmarking Best Practices

### Proper Benchmark Structure

```go
func BenchmarkOperation(b *testing.B) {
    // Setup outside the loop
    data := prepareData()

    b.ResetTimer() // Reset after setup

    for i := 0; i < b.N; i++ {
        operation(data)
    }
}
```

### Memory Benchmarks

```bash
go test -bench=. -benchmem
```

```go
// Results show:
// BenchmarkX    1000000    1234 ns/op    256 B/op    4 allocs/op
//                                        ^^^^^^^^    ^^^^^^^^^^^^
//                                        bytes/op    allocations/op
```

### Profile Before Optimizing

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Trace
go test -trace=trace.out -bench=.
go tool trace trace.out
```

---

## Quick Wins

1. **Pre-allocate slices and maps** when size is known
2. **Use strings.Builder** for string concatenation
3. **Use sync.Pool** for frequently allocated objects
4. **Avoid interface{}** in hot paths
5. **Buffer I/O** operations
6. **Profile before optimizing** - don't guess
7. **Use -benchmem** to track allocations
8. **Order struct fields** by size for better alignment
9. **Return small structs by value**, not pointer
10. **Use worker pools** to limit goroutine explosion
