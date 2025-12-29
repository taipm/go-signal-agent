# HTTP Patterns

> "Talk is cheap. Show me the code." — Linus Torvalds

---

## TL;DR - Server Template

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    server := &http.Server{
        Addr:         ":8080",
        Handler:      newRouter(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    // Wait for shutdown
    <-ctx.Done()
    log.Println("Shutting down...")

    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

---

## Server Patterns

### ❌ BROKEN: No Timeouts

```go
// 🔴 THẢM HỌA — Slowloris attack sẽ KILL server
server := &http.Server{
    Addr:    ":8080",
    Handler: handler,
    // NO TIMEOUTS = Connections hang forever
}
```

**Tại sao chết:**
- Attacker mở connection, gửi data chậm
- Server giữ connection mãi mãi
- File descriptors exhausted → server dead

### ✅ ĐÚNG: Proper Timeouts

```go
server := &http.Server{
    Addr:    ":8080",
    Handler: handler,

    // Read timeout: time to read request headers + body
    ReadTimeout: 15 * time.Second,

    // Write timeout: time to write response
    WriteTimeout: 15 * time.Second,

    // Idle timeout: keep-alive connections
    IdleTimeout: 60 * time.Second,

    // Header timeout: time to read request headers only
    ReadHeaderTimeout: 5 * time.Second,

    // Max header size
    MaxHeaderBytes: 1 << 20, // 1 MB
}
```

---

### ❌ BROKEN: No Graceful Shutdown

```go
// 🔴 THẢM HỌA — In-flight requests DROPPED
func main() {
    http.ListenAndServe(":8080", handler)
    // Ctrl+C = immediate death
    // Requests in progress = LOST
}
```

### ✅ ĐÚNG: Graceful Shutdown

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    server := &http.Server{Addr: ":8080", Handler: handler}

    // Start in goroutine
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    log.Println("Server started on :8080")

    // Wait for signal
    <-ctx.Done()
    log.Println("Shutting down...")

    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }

    log.Println("Server stopped")
}
```

---

## Handler Patterns

### ❌ BROKEN: Panic Crashes Server

```go
// 🔴 THẢM HỌA — One panic = entire server DOWN
func handler(w http.ResponseWriter, r *http.Request) {
    data := riskyOperation()  // May panic
    json.NewEncoder(w).Encode(data)
}
```

### ✅ ĐÚNG: Panic Recovery Middleware

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Log with stack trace
                log.Printf("PANIC: %v\n%s", err, debug.Stack())

                // Return 500
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "internal server error",
                })
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// Usage
router := http.NewServeMux()
router.HandleFunc("/api/data", dataHandler)
server := &http.Server{
    Handler: recoveryMiddleware(router),
}
```

---

### ❌ BROKEN: No Request Size Limit

```go
// 🔴 THẢM HỌA — Attacker uploads 10GB file = OOM
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)  // Memory bomb!
    process(body)
}
```

### ✅ ĐÚNG: Request Size Limits

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Limit request body to 10MB
    r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

    // Parse with limit
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
        return
    }

    file, _, err := r.FormFile("upload")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Stream to disk, don't load in memory
    dst, err := os.Create("/tmp/upload")
    if err != nil {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    // Copy with size limit
    if _, err := io.CopyN(dst, file, 10<<20); err != nil && err != io.EOF {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }
}
```

---

## Context Propagation

### ❌ BROKEN: Ignoring Request Context

```go
// 🔴 BROKEN — Client disconnect = wasted resources
func handler(w http.ResponseWriter, r *http.Request) {
    result := expensiveQuery()  // Runs even if client gone
    json.NewEncoder(w).Encode(result)
}
```

### ✅ ĐÚNG: Use Request Context

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Pass context to all operations
    result, err := expensiveQuery(ctx)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            // Client disconnected, stop processing
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}

func expensiveQuery(ctx context.Context) (*Result, error) {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Use context in database calls
    rows, err := db.QueryContext(ctx, "SELECT ...")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // ...
}
```

---

## Client Patterns

### ❌ BROKEN: Default http.Client

```go
// 🔴 BROKEN — No timeout, no connection pooling control
resp, err := http.Get("https://api.example.com/data")
```

### ✅ ĐÚNG: Configured Client

```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        // Connection pooling
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,

        // Timeouts
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}

func fetchData(ctx context.Context, url string) (*Data, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    resp, err := httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    // Limit response body
    body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
    if err != nil {
        return nil, fmt.Errorf("read body: %w", err)
    }

    var data Data
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, fmt.Errorf("unmarshal: %w", err)
    }

    return &data, nil
}
```

---

## Retry Pattern

```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

func fetchWithRetry(ctx context.Context, url string, cfg RetryConfig) (*http.Response, error) {
    var lastErr error

    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        // Check context
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, err
        }

        resp, err := httpClient.Do(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }

        if resp != nil {
            resp.Body.Close()
        }
        lastErr = err

        // Exponential backoff
        delay := cfg.BaseDelay * time.Duration(1<<attempt)
        if delay > cfg.MaxDelay {
            delay = cfg.MaxDelay
        }

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
        }
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

## Middleware Stack

```go
// Middleware type
type Middleware func(http.Handler) http.Handler

// Chain middlewares
func Chain(middlewares ...Middleware) Middleware {
    return func(final http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            final = middlewares[i](final)
        }
        return final
    }
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, status: 200}

        next.ServeHTTP(wrapped, r)

        slog.Info("request",
            slog.String("method", r.Method),
            slog.String("path", r.URL.Path),
            slog.Int("status", wrapped.status),
            slog.Duration("duration", time.Since(start)),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

// Request ID middleware
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        ctx := context.WithValue(r.Context(), "request_id", requestID)
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Usage
handler := Chain(
    recoveryMiddleware,
    loggingMiddleware,
    requestIDMiddleware,
)(router)
```

---

## JSON Response Helpers

```go
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("JSON encode error: %v", err)
    }
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}

// Usage
func userHandler(w http.ResponseWriter, r *http.Request) {
    user, err := getUser(r.Context(), r.URL.Query().Get("id"))
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            respondError(w, http.StatusNotFound, "user not found")
            return
        }
        respondError(w, http.StatusInternalServerError, "internal error")
        return
    }

    respondJSON(w, http.StatusOK, user)
}
```

---

## Checklist

- [ ] Server has ReadTimeout, WriteTimeout, IdleTimeout
- [ ] Graceful shutdown với context
- [ ] Panic recovery middleware
- [ ] Request body size limits (http.MaxBytesReader)
- [ ] Context propagation từ request
- [ ] Client có custom Transport với timeouts
- [ ] Retry pattern cho external calls
- [ ] Logging middleware
- [ ] Request ID tracking

---

## Quick Reference

| Pattern | Purpose |
|---------|---------|
| ReadTimeout | Prevent slow readers |
| WriteTimeout | Prevent slow writers |
| MaxBytesReader | Prevent memory bombs |
| http.MaxBytesReader | Limit upload size |
| server.Shutdown(ctx) | Graceful shutdown |
| r.Context() | Request cancellation |
| http.NewRequestWithContext | Client context |
| Custom Transport | Connection pooling |

---

**Talk is cheap. Show me the code.**
