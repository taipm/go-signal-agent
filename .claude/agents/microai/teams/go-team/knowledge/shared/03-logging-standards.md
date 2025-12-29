# Logging Standards - Shared Knowledge

**Version:** 1.0.0
**Applies to:** All Agents

---

## TL;DR

- Dùng structured logging (slog hoặc zerolog)
- Log levels: Debug, Info, Warn, Error
- Include context fields (request_id, user_id)
- Không log sensitive data
- Log errors chỉ ở top level

---

## 1. Supported Libraries

### slog (Go 1.21+ - Recommended)

```go
import "log/slog"

// Setup
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Usage
slog.Info("user created",
    slog.String("user_id", user.ID),
    slog.String("email", user.Email),
)

slog.Error("failed to create user",
    slog.String("email", email),
    slog.Any("error", err),
)
```

### zerolog (Alternative)

```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// Setup
zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

// Usage
log.Info().
    Str("user_id", user.ID).
    Str("email", user.Email).
    Msg("user created")

log.Error().
    Err(err).
    Str("email", email).
    Msg("failed to create user")
```

---

## 2. Log Levels

### Level Guidelines

| Level | When to Use | Example |
|-------|------------|---------|
| Debug | Detailed debugging info | Variable values, function entry/exit |
| Info | Normal operations | User created, request processed |
| Warn | Potential issues | Retry attempt, degraded service |
| Error | Actual failures | Database error, external API failure |

### Examples

```go
// ✅ Debug - detailed tracing
slog.Debug("processing item",
    slog.String("item_id", item.ID),
    slog.Int("retry", retryCount),
)

// ✅ Info - business events
slog.Info("order placed",
    slog.String("order_id", order.ID),
    slog.String("user_id", order.UserID),
    slog.Float64("total", order.Total),
)

// ✅ Warn - recoverable issues
slog.Warn("rate limit approaching",
    slog.String("client_id", clientID),
    slog.Int("remaining", remaining),
)

// ✅ Error - failures
slog.Error("payment failed",
    slog.String("order_id", orderID),
    slog.Any("error", err),
)
```

---

## 3. Context Fields

### Request Context

```go
// ✅ CORRECT - include trace context
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-ID")
    userID := getUserID(r.Context())

    logger := slog.With(
        slog.String("request_id", requestID),
        slog.String("user_id", userID),
        slog.String("path", r.URL.Path),
        slog.String("method", r.Method),
    )

    logger.Info("request started")

    // Pass logger in context
    ctx := WithLogger(r.Context(), logger)

    // ... handle request

    logger.Info("request completed",
        slog.Int("status", status),
        slog.Duration("duration", time.Since(start)),
    )
}
```

### Logger in Context

```go
type contextKey string

const loggerKey contextKey = "logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}

// Usage
func (s *Service) DoWork(ctx context.Context) error {
    logger := LoggerFrom(ctx)
    logger.Info("starting work")
    // ...
}
```

---

## 4. Standard Fields

### Required Fields

```go
// ✅ CORRECT - consistent field names
slog.Info("event name",
    slog.String("request_id", requestID),    // trace correlation
    slog.String("user_id", userID),          // who
    slog.String("action", "create_order"),   // what
    slog.String("resource", "order"),        // which resource
    slog.String("resource_id", orderID),     // resource identifier
)
```

### Error Logging

```go
// ✅ CORRECT - error with stack
slog.Error("operation failed",
    slog.String("operation", "get_user"),
    slog.String("user_id", userID),
    slog.Any("error", err),
)

// zerolog version
log.Error().
    Err(err).
    Str("operation", "get_user").
    Str("user_id", userID).
    Msg("operation failed")
```

### Performance Metrics

```go
// ✅ CORRECT - duration logging
start := time.Now()
result, err := expensiveOperation(ctx)
duration := time.Since(start)

slog.Info("operation completed",
    slog.String("operation", "expensive_operation"),
    slog.Duration("duration", duration),
    slog.Int("result_count", len(result)),
)
```

---

## 5. Sensitive Data

### Never Log These

```go
// ❌ WRONG - logging sensitive data
slog.Info("user login",
    slog.String("password", password),        // NEVER
    slog.String("api_key", apiKey),           // NEVER
    slog.String("credit_card", cardNumber),   // NEVER
    slog.String("ssn", socialSecurityNumber), // NEVER
    slog.String("token", authToken),          // NEVER
)

// ✅ CORRECT - mask or omit
slog.Info("user login",
    slog.String("email", email),
    slog.Bool("has_password", password != ""),
)

slog.Info("payment processed",
    slog.String("card_last_four", cardNumber[len(cardNumber)-4:]),
)
```

### Masking Helper

```go
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    if len(parts[0]) <= 2 {
        return "***@" + parts[1]
    }
    return parts[0][:2] + "***@" + parts[1]
}

func MaskCard(card string) string {
    if len(card) < 4 {
        return "****"
    }
    return "****" + card[len(card)-4:]
}

// Usage
slog.Info("payment",
    slog.String("email", MaskEmail(user.Email)),
    slog.String("card", MaskCard(cardNumber)),
)
```

---

## 6. HTTP Request Logging

### Middleware

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        // Add logger to context
        logger := slog.With(
            slog.String("request_id", requestID),
            slog.String("method", r.Method),
            slog.String("path", r.URL.Path),
            slog.String("remote_addr", r.RemoteAddr),
        )

        ctx := WithLogger(r.Context(), logger)
        r = r.WithContext(ctx)

        // Set request ID header
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(wrapped, r)

        logger.Info("request completed",
            slog.Int("status", wrapped.status),
            slog.Duration("duration", time.Since(start)),
            slog.Int64("bytes", wrapped.bytes),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
    bytes  int64
}

func (w *responseWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
    n, err := w.ResponseWriter.Write(b)
    w.bytes += int64(n)
    return n, err
}
```

---

## 7. Error Logging Rules

### Only Log at Top Level

```go
// ✅ CORRECT - return errors, log at handler
func (r *Repo) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := r.db.Get(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("query user %s: %w", id, err)
    }
    return user, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    return user, nil
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        // Log here at the top level
        LoggerFrom(r.Context()).Error("get user failed",
            slog.String("user_id", id),
            slog.Any("error", err),
        )
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

### ❌ Don't Log and Return

```go
// ❌ WRONG - double logging
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        slog.Error("failed to get user", slog.Any("error", err))  // DON'T
        return nil, err  // caller will also log
    }
    return user, nil
}
```

---

## 8. JSON Output Format

```json
{
  "time": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "msg": "order placed",
  "request_id": "abc-123",
  "user_id": "user-456",
  "order_id": "order-789",
  "total": 99.99,
  "items_count": 3
}
```

---

## Quick Reference

| Task | slog | zerolog |
|------|------|---------|
| Info log | `slog.Info("msg", slog.String("k", "v"))` | `log.Info().Str("k", "v").Msg("msg")` |
| Error log | `slog.Error("msg", slog.Any("error", err))` | `log.Error().Err(err).Msg("msg")` |
| With context | `slog.With(slog.String("k", "v"))` | `log.With().Str("k", "v").Logger()` |
| Duration | `slog.Duration("d", time.Second)` | `.Dur("d", time.Second)` |

---

## Related Knowledge

- [01-go-fundamentals.md](./01-go-fundamentals.md) - Basic patterns
- [02-error-patterns.md](./02-error-patterns.md) - Error handling
