# Security Vulnerability Patterns

## ALL Security Issues = 🔴 BROKEN

Security is non-negotiable. Every security issue is critical.

---

## SQL Injection

### Vulnerable Patterns

```go
// 🔴 BROKEN - String concatenation
query := "SELECT * FROM users WHERE id = " + userID
db.Query(query)

// 🔴 BROKEN - fmt.Sprintf
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", userName)
db.Query(query)

// 🔴 BROKEN - String interpolation
db.Query("DELETE FROM users WHERE id = " + strconv.Itoa(id))
```

### Safe Patterns

```go
// OK - Parameterized queries
db.Query("SELECT * FROM users WHERE id = $1", userID)

// OK - Prepared statements
stmt, _ := db.Prepare("SELECT * FROM users WHERE name = ?")
stmt.Query(userName)

// OK - Query builders (sqlx, squirrel)
sq.Select("*").From("users").Where(sq.Eq{"id": userID})
```

### Detection Regex

```regex
db\.(Query|Exec)\([^,]*\+
fmt\.Sprintf\([^)]*SELECT
fmt\.Sprintf\([^)]*INSERT
fmt\.Sprintf\([^)]*UPDATE
fmt\.Sprintf\([^)]*DELETE
```

---

## Command Injection

### Vulnerable Patterns

```go
// 🔴 BROKEN - User input in command
cmd := exec.Command("bash", "-c", "ls " + userInput)

// 🔴 BROKEN - Shell execution
cmd := exec.Command("sh", "-c", userCommand)

// 🔴 BROKEN - Unvalidated input
exec.Command(userProvidedProgram, args...)
```

### Safe Patterns

```go
// OK - Separate arguments (no shell)
cmd := exec.Command("ls", "-la", directory)

// OK - Whitelist validation
allowedCommands := map[string]bool{"ls": true, "cat": true}
if !allowedCommands[command] {
    return ErrInvalidCommand
}

// OK - Escape if shell is necessary
import "github.com/alessio/shellescape"
cmd := exec.Command("bash", "-c", "echo " + shellescape.Quote(userInput))
```

### Detection Regex

```regex
exec\.Command\("(bash|sh|cmd)",\s*"-c"
exec\.Command\([^,]*\+
os\.System\(
```

---

## Path Traversal

### Vulnerable Patterns

```go
// 🔴 BROKEN - Direct user input in path
filepath := "/var/data/" + userInput
ioutil.ReadFile(filepath)

// 🔴 BROKEN - Unvalidated relative paths
http.ServeFile(w, r, r.URL.Path)

// 🔴 BROKEN - No sanitization
os.Open(basePath + "/" + filename)
```

### Safe Patterns

```go
// OK - Clean and validate
cleanPath := filepath.Clean(userInput)
if strings.Contains(cleanPath, "..") {
    return ErrInvalidPath
}

// OK - Ensure within base directory
absPath := filepath.Join(basePath, filepath.Clean(userInput))
if !strings.HasPrefix(absPath, basePath) {
    return ErrPathTraversal
}

// OK - Use http.Dir with restrictions
http.FileServer(http.Dir("/safe/directory"))
```

### Detection Regex

```regex
filepath\.Join\([^,]+,\s*[^)]*userInput
ioutil\.ReadFile\([^)]*\+
os\.Open\([^)]*\+
```

---

## XSS (Cross-Site Scripting)

### Vulnerable Patterns

```go
// 🔴 BROKEN - Unescaped HTML output
fmt.Fprintf(w, "<h1>%s</h1>", userInput)

// 🔴 BROKEN - template.HTML bypass
tmpl.Execute(w, template.HTML(userInput))

// 🔴 BROKEN - Raw JSON in HTML
fmt.Fprintf(w, "<script>var data = %s;</script>", jsonData)
```

### Safe Patterns

```go
// OK - html/template (auto-escapes)
tmpl := template.Must(template.New("page").Parse(`<h1>{{.Title}}</h1>`))
tmpl.Execute(w, data)

// OK - Explicit escaping
html.EscapeString(userInput)

// OK - JSON encoding for scripts
json.Marshal(data) // Then use in template
```

### Detection Regex

```regex
fmt\.Fprintf\(w,.*<.*%s
template\.HTML\(
\.Write\(\[\]byte\(.*\+
```

---

## Hardcoded Secrets

### Vulnerable Patterns

```go
// 🔴 BROKEN - Hardcoded credentials
const password = "admin123"
apiKey := "sk-1234567890abcdef"

// 🔴 BROKEN - In connection strings
db, _ := sql.Open("postgres", "user:password@localhost/db")

// 🔴 BROKEN - JWT secrets
jwtSecret := []byte("my-secret-key")
```

### Safe Patterns

```go
// OK - Environment variables
password := os.Getenv("DB_PASSWORD")

// OK - Secret management
secret, _ := vault.GetSecret("api-key")

// OK - Configuration files (not in repo)
config, _ := loadConfig("/etc/app/config.yaml")
```

---

## Insecure Randomness

### Vulnerable Patterns

```go
// 🔴 BROKEN - math/rand for security
import "math/rand"
token := rand.Intn(1000000)
sessionID := fmt.Sprintf("%d", rand.Int63())
```

### Safe Patterns

```go
// OK - crypto/rand for security
import "crypto/rand"
bytes := make([]byte, 32)
crypto_rand.Read(bytes)

// OK - Use secure token generators
import "github.com/google/uuid"
token := uuid.New().String()
```

### Detection Regex

```regex
math/rand
rand\.(Int|Intn|Int63|Float)\(
rand\.Seed\(
```

---

## Weak Cryptography

### Vulnerable Patterns

```go
// 🔴 BROKEN - MD5 for passwords
hash := md5.Sum([]byte(password))

// 🔴 BROKEN - SHA1 for security
hash := sha1.Sum([]byte(data))

// 🔴 BROKEN - DES encryption
cipher.NewDESCipher(key)

// 🔴 BROKEN - ECB mode
cipher.NewCBCEncrypter(block, iv) // Check if IV is reused
```

### Safe Patterns

```go
// OK - bcrypt for passwords
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// OK - SHA-256 or better
hash := sha256.Sum256(data)

// OK - AES-GCM
block, _ := aes.NewCipher(key)
gcm, _ := cipher.NewGCM(block)
```

### Detection Regex

```regex
md5\.
sha1\.
des\.
"crypto/des"
"crypto/md5"
```

---

## Unsafe Pointer Usage

### Vulnerable Patterns

```go
// 🔴 BROKEN - Unsafe memory access
import "unsafe"
ptr := unsafe.Pointer(&data)
*(*int)(ptr) = value

// 🔴 BROKEN - reflect.SliceHeader manipulation
header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
```

### Safe Patterns

```go
// OK - Avoid unsafe unless absolutely necessary
// If needed, document thoroughly and isolate

// OK - Use type-safe alternatives
binary.Read(reader, binary.LittleEndian, &data)
```

---

## Race Conditions (Security Context)

### Vulnerable Patterns

```go
// 🔴 BROKEN - TOCTOU (Time of Check to Time of Use)
if fileExists(path) {  // Check
    data, _ := ioutil.ReadFile(path)  // Use - file might have changed
}

// 🔴 BROKEN - Auth check followed by action
if user.IsAdmin() {
    // Another goroutine might change user.role here
    performAdminAction()
}
```

### Safe Patterns

```go
// OK - Atomic operations
file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0600)

// OK - Hold lock during check and use
mu.Lock()
if user.IsAdmin() {
    performAdminAction()
}
mu.Unlock()
```

---

## HTTP Security Headers

### Missing Headers (🟡 SMELL in code, 🔴 in production)

```go
// Should add security headers
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("X-XSS-Protection", "1; mode=block")
w.Header().Set("Content-Security-Policy", "default-src 'self'")
w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

---

## TLS/SSL Issues

### Vulnerable Patterns

```go
// 🔴 BROKEN - Skip TLS verification
&http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// 🔴 BROKEN - Weak TLS version
&tls.Config{MinVersion: tls.VersionTLS10}
```

### Safe Patterns

```go
// OK - Proper TLS config
&tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    },
}
```
