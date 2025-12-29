# OWASP Top 10 - Security Agent Knowledge

**Version:** 1.0.0
**Agent:** Security Agent

---

## TL;DR

- SQL Injection: dùng parameterized queries
- XSS: escape output, Content-Type headers
- Auth: bcrypt passwords, secure sessions
- Secrets: env vars, không hardcode
- Input: validate tất cả user input

---

## 1. A01: Broken Access Control

### Vulnerable Code

```go
// ❌ WRONG - No authorization check
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    orderID := chi.URLParam(r, "id")
    order, _ := h.service.GetOrder(r.Context(), orderID)
    // Anyone can access any order!
    json.NewEncoder(w).Encode(order)
}

// ❌ WRONG - Client-side only check
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    // Trusting frontend to send correct user
    userID := r.Header.Get("X-User-ID")
    h.service.DeleteUser(r.Context(), userID)
}
```

### Secure Code

```go
// ✅ CORRECT - Server-side authorization
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    orderID := chi.URLParam(r, "id")
    userID := GetUserIDFromContext(r.Context())

    order, err := h.service.GetOrder(r.Context(), orderID)
    if err != nil {
        respondError(w, http.StatusNotFound, "order not found")
        return
    }

    // Check ownership
    if order.UserID != userID {
        h.logger.Warn("unauthorized access attempt",
            slog.String("user_id", userID),
            slog.String("order_id", orderID),
        )
        respondError(w, http.StatusForbidden, "access denied")
        return
    }

    json.NewEncoder(w).Encode(order)
}

// ✅ CORRECT - Role-based access control
func AdminOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r.Context())
        if user == nil || user.Role != RoleAdmin {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## 2. A02: Cryptographic Failures

### Vulnerable Code

```go
// ❌ WRONG - Weak hashing
import "crypto/md5"
func hashPassword(password string) string {
    return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

// ❌ WRONG - Hardcoded secrets
const jwtSecret = "my-secret-key"

// ❌ WRONG - Weak encryption
import "crypto/des"
```

### Secure Code

```go
// ✅ CORRECT - bcrypt for passwords
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// ✅ CORRECT - Secrets from environment
func GetJWTSecret() []byte {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET not set")
    }
    if len(secret) < 32 {
        log.Fatal("JWT_SECRET must be at least 32 characters")
    }
    return []byte(secret)
}

// ✅ CORRECT - Strong encryption (AES-256-GCM)
import "crypto/aes"
import "crypto/cipher"

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
```

---

## 3. A03: Injection

### SQL Injection

```go
// ❌ WRONG - String concatenation
func (r *Repo) GetUser(name string) (*User, error) {
    query := "SELECT * FROM users WHERE name = '" + name + "'"
    // Input: "'; DROP TABLE users; --"
    return r.db.Query(query)
}

// ❌ WRONG - fmt.Sprintf
func (r *Repo) GetUser(id string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)
    return r.db.Query(query)
}
```

```go
// ✅ CORRECT - Parameterized queries
func (r *Repo) GetUser(ctx context.Context, id string) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    row := r.db.QueryRowContext(ctx, query, id)

    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// ✅ CORRECT - Using ORM
func (r *Repo) GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
    return &user, result.Error
}
```

### Command Injection

```go
// ❌ WRONG - Unsanitized input
func RunCommand(filename string) error {
    cmd := exec.Command("sh", "-c", "cat "+filename)
    // Input: "; rm -rf /"
    return cmd.Run()
}
```

```go
// ✅ CORRECT - Avoid shell, use arguments
func RunCommand(filename string) error {
    // Validate filename
    if strings.ContainsAny(filename, ";&|`$") {
        return errors.New("invalid filename")
    }

    cmd := exec.Command("cat", filename)  // No shell
    return cmd.Run()
}

// ✅ BETTER - Avoid exec entirely
func ReadFile(filename string) ([]byte, error) {
    // Validate path
    cleanPath := filepath.Clean(filename)
    if strings.HasPrefix(cleanPath, "..") {
        return nil, errors.New("invalid path")
    }

    return os.ReadFile(cleanPath)
}
```

---

## 4. A04: Insecure Design

### Rate Limiting

```go
// ✅ CORRECT - Rate limiting
import "golang.org/x/time/rate"

type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
}

func (r *RateLimiter) GetLimiter(ip string) *rate.Limiter {
    r.mu.Lock()
    defer r.mu.Unlock()

    limiter, exists := r.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(10, 30)  // 10 req/sec, burst 30
        r.visitors[ip] = limiter
    }
    return limiter
}

func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            limiter := rl.GetLimiter(ip)

            if !limiter.Allow() {
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 5. A05: Security Misconfiguration

### Secure Headers

```go
// ✅ CORRECT - Security headers middleware
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent XSS
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Content Security Policy
        w.Header().Set("Content-Security-Policy", "default-src 'self'")

        // HTTPS only
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        next.ServeHTTP(w, r)
    })
}
```

### CORS Configuration

```go
// ✅ CORRECT - Restrictive CORS
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
    allowed := make(map[string]bool)
    for _, origin := range allowedOrigins {
        allowed[origin] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            if allowed[origin] {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "86400")
            }

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// ❌ WRONG - Allow all origins
w.Header().Set("Access-Control-Allow-Origin", "*")
```

---

## 6. A06: Vulnerable Components

### Dependency Scanning

```bash
# Check for vulnerabilities
go list -m -json all | nancy sleuth

# Using govulncheck (official Go tool)
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Update dependencies
go get -u ./...
go mod tidy
```

---

## 7. A07: Auth Failures

### Session Management

```go
// ✅ CORRECT - Secure session
func CreateSession(w http.ResponseWriter, userID string) error {
    sessionID := generateSecureToken(32)

    // Store session server-side
    if err := sessionStore.Set(sessionID, userID, 24*time.Hour); err != nil {
        return err
    }

    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Path:     "/",
        HttpOnly: true,      // No JavaScript access
        Secure:   true,      // HTTPS only
        SameSite: http.SameSiteStrictMode,
        MaxAge:   86400,     // 24 hours
    })

    return nil
}

func generateSecureToken(length int) string {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        panic(err)
    }
    return base64.URLEncoding.EncodeToString(bytes)
}
```

### JWT Best Practices

```go
// ✅ CORRECT - Secure JWT
func GenerateToken(userID string) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "iat": time.Now().Unix(),
        "exp": time.Now().Add(15 * time.Minute).Unix(),  // Short expiry
        "jti": uuid.New().String(),  // Unique token ID
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(GetJWTSecret())
}

func ValidateToken(tokenString string) (*jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return GetJWTSecret(), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &claims, nil
    }

    return nil, errors.New("invalid token")
}
```

---

## 8. A08: Data Integrity Failures

### Input Validation

```go
import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"gte=0,lte=150"`
}

var validate = validator.New()

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    if err := validate.Struct(req); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Proceed with validated data
}
```

---

## 9. A09: Logging & Monitoring

### Security Logging

```go
// ✅ CORRECT - Security event logging
func LogSecurityEvent(ctx context.Context, event string, details map[string]interface{}) {
    logger := slog.Default()

    attrs := []slog.Attr{
        slog.String("event_type", "security"),
        slog.String("event", event),
        slog.Time("timestamp", time.Now()),
    }

    // Add request context
    if requestID, ok := ctx.Value("request_id").(string); ok {
        attrs = append(attrs, slog.String("request_id", requestID))
    }

    // Add details
    for k, v := range details {
        attrs = append(attrs, slog.Any(k, v))
    }

    logger.LogAttrs(ctx, slog.LevelWarn, "security event", attrs...)
}

// Usage
LogSecurityEvent(ctx, "failed_login", map[string]interface{}{
    "email":      email,
    "ip":         r.RemoteAddr,
    "user_agent": r.UserAgent(),
    "attempts":   failedAttempts,
})
```

---

## 10. A10: SSRF

### Server-Side Request Forgery Prevention

```go
// ❌ WRONG - User-controlled URL
func FetchURL(w http.ResponseWriter, r *http.Request) {
    url := r.URL.Query().Get("url")
    resp, _ := http.Get(url)  // Can access internal services!
    io.Copy(w, resp.Body)
}
```

```go
// ✅ CORRECT - URL validation
var allowedHosts = map[string]bool{
    "api.example.com":     true,
    "cdn.example.com":     true,
}

func FetchURL(w http.ResponseWriter, r *http.Request) {
    urlStr := r.URL.Query().Get("url")

    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "invalid URL")
        return
    }

    // Validate scheme
    if parsedURL.Scheme != "https" {
        respondError(w, http.StatusBadRequest, "only HTTPS allowed")
        return
    }

    // Validate host
    if !allowedHosts[parsedURL.Host] {
        respondError(w, http.StatusBadRequest, "host not allowed")
        return
    }

    // Block private IPs
    ips, err := net.LookupIP(parsedURL.Hostname())
    if err != nil {
        respondError(w, http.StatusBadRequest, "cannot resolve host")
        return
    }

    for _, ip := range ips {
        if isPrivateIP(ip) {
            respondError(w, http.StatusBadRequest, "private IP not allowed")
            return
        }
    }

    // Safe to fetch
    resp, err := http.Get(urlStr)
    // ...
}

func isPrivateIP(ip net.IP) bool {
    private := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "127.0.0.0/8",
        "169.254.0.0/16",
    }
    for _, cidr := range private {
        _, network, _ := net.ParseCIDR(cidr)
        if network.Contains(ip) {
            return true
        }
    }
    return false
}
```

---

## Quick Reference

| Vulnerability | Fix |
|--------------|-----|
| SQL Injection | Parameterized queries |
| XSS | Escape output, CSP headers |
| Broken Auth | bcrypt, secure sessions |
| Secrets | Environment variables |
| SSRF | URL whitelist, block private IPs |

---

## Related Knowledge

- [02-secure-coding.md](./02-secure-coding.md) - Secure coding patterns
- [../shared/02-error-patterns.md](../shared/02-error-patterns.md) - Error handling
