# Hardcode Detection Patterns

## Severity Classification

| Pattern Type | Severity | Reason |
|-------------|----------|--------|
| Credentials | 🔴 BROKEN | Security disaster |
| API Keys | 🔴 BROKEN | Security disaster |
| Secrets/Tokens | 🔴 BROKEN | Security disaster |
| Database URLs | 🔴 BROKEN | Security + config issue |
| Magic Numbers | 🟡 SMELL | Maintainability |
| Hardcoded URLs | 🟡 SMELL | Config issue |
| Hardcoded Paths | 🟡 SMELL | Portability issue |
| String Literals | 🟡 SMELL | i18n/maintainability |

## Regex Patterns for Detection

### Credentials (🔴 BROKEN)

```regex
# Passwords
password\s*[:=]\s*["'][^"']+["']
passwd\s*[:=]\s*["'][^"']+["']
pwd\s*[:=]\s*["'][^"']+["']
secret\s*[:=]\s*["'][^"']+["']

# In Go specifically
Password\s*[:=]\s*"[^"]+"
password\s*:\s*"[^"]+"
```

### API Keys (🔴 BROKEN)

```regex
# Generic API keys
(api[_-]?key|apikey)\s*[:=]\s*["'][^"']+["']
(api[_-]?secret)\s*[:=]\s*["'][^"']+["']

# Specific services
AKIA[0-9A-Z]{16}                    # AWS Access Key
AIza[0-9A-Za-z\-_]{35}              # Google API Key
sk-[a-zA-Z0-9]{48}                  # OpenAI API Key
ghp_[a-zA-Z0-9]{36}                 # GitHub Personal Access Token
glpat-[a-zA-Z0-9\-]{20}             # GitLab Personal Access Token
xox[baprs]-[0-9a-zA-Z]{10,48}       # Slack Token
```

### Tokens & Secrets (🔴 BROKEN)

```regex
# JWT (not always hardcoded, check context)
eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*

# Bearer tokens
Bearer\s+[a-zA-Z0-9\-_]+

# Generic secrets
(secret|token|auth)\s*[:=]\s*["'][^"']{8,}["']
```

### Database Connection Strings (🔴 BROKEN)

```regex
# PostgreSQL
postgres(ql)?://[^:]+:[^@]+@

# MySQL
mysql://[^:]+:[^@]+@

# MongoDB
mongodb(\+srv)?://[^:]+:[^@]+@

# Redis with password
redis://:[^@]+@

# Generic DSN
(user|username)=[^&\s]+.*(password|passwd|pwd)=[^&\s]+
```

### URLs (🟡 SMELL)

```regex
# HTTP/HTTPS URLs (check if should be configurable)
https?://[^\s"']+

# Localhost URLs (often OK in dev, bad in prod)
https?://localhost(:\d+)?
https?://127\.0\.0\.1(:\d+)?
```

### IP Addresses (🟡 SMELL)

```regex
# IPv4
\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b

# With port
\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b
```

### Ports (🟡 SMELL)

```regex
# Standalone port numbers (context-dependent)
:\d{4,5}\b

# Port assignments
port\s*[:=]\s*\d{4,5}
Port\s*[:=]\s*\d{4,5}
```

### Magic Numbers (🟡 SMELL)

```regex
# Large unexplained numbers
\b\d{4,}\b

# Common problematic patterns
time\.Sleep\(\d+\s*\*
make\(\[\]\w+,\s*0,\s*\d{3,}\)
timeout\s*[:=]\s*\d+

# But NOT these (acceptable):
# - 0, 1, 2 (often OK)
# - 100, 1000 (often OK for percentages/multipliers)
# - Known constants like 1024, 4096
```

## Go-Specific Patterns

### Common Hardcode Locations

```go
// In struct definitions
type Config struct {
    Host     string // Look for default values
    Port     int    // Magic numbers
    Timeout  int    // Magic numbers
    APIKey   string // Should NEVER have default
}

// In init() functions
func init() {
    // Often contains hardcoded config
}

// In const blocks
const (
    defaultTimeout = 30  // Is 30 explained?
    maxRetries = 5       // Why 5?
)

// String literals in function calls
http.Get("https://api.example.com/v1/users")
sql.Open("postgres", "host=localhost user=admin password=secret")
```

### Environment Variable Checks

```go
// SMELL - Default to hardcoded value
host := os.Getenv("HOST")
if host == "" {
    host = "localhost"  // Should this be configurable elsewhere?
}

// Better - Fail if not configured
host := os.Getenv("HOST")
if host == "" {
    log.Fatal("HOST environment variable not set")
}
```

## Grep Commands for Detection

```bash
# Credentials
grep -rn "password\s*[:=]\s*\"" --include="*.go"
grep -rn "secret\s*[:=]\s*\"" --include="*.go"

# API Keys
grep -rn "apikey\|api_key\|api-key" --include="*.go"
grep -rn "AKIA[0-9A-Z]" --include="*.go"

# URLs
grep -rn "https\?://" --include="*.go"

# Magic numbers (be selective)
grep -rn "time\.Sleep([0-9]" --include="*.go"
```

## False Positives to Ignore

- Test files with mock data
- Example/documentation strings
- Error messages containing URLs
- Version strings
- Well-known constants (http.StatusOK = 200)
- Format strings with %d, %s etc.

## Recommendations

When hardcoded values are found:

1. **Credentials**: Move to environment variables or secret manager IMMEDIATELY
2. **API Keys**: Use secret management (Vault, AWS Secrets Manager)
3. **URLs**: Use configuration files or environment variables
4. **Magic Numbers**: Define as named constants with documentation
5. **Timeouts**: Use configuration with sensible defaults documented
