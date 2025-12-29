# Pattern Detector

**Version:** 1.0.0

Detection rules cho các patterns phổ biến trong Go codebases.

---

## Detection Categories

### 1. Architecture Patterns

#### Clean Architecture

```yaml
clean_architecture:
  confidence_threshold: 0.7

  indicators:
    strong:  # +0.3 each
      - directory "internal/domain" exists
      - directory "internal/usecase" exists
      - directory "internal/entity" exists
      - interfaces in domain package

    medium:  # +0.2 each
      - directory "internal/repository" exists
      - directory "internal/delivery" exists
      - clear layer separation

    weak:  # +0.1 each
      - no direct DB calls in handlers
      - dependency injection used

  detection_commands:
    - "find . -type d -name 'domain'"
    - "find . -type d -name 'usecase'"
    - "grep -r 'type.*Repository interface' internal/"
```

#### Hexagonal Architecture

```yaml
hexagonal:
  confidence_threshold: 0.7

  indicators:
    strong:
      - directory "internal/ports" exists
      - directory "internal/adapters" exists
      - "ports" contains interfaces only

    medium:
      - "adapters/primary" and "adapters/secondary"
      - clear port definitions

  detection_commands:
    - "find . -type d -name 'ports'"
    - "find . -type d -name 'adapters'"
```

#### Simple Layered

```yaml
simple_layered:
  confidence_threshold: 0.6

  indicators:
    strong:
      - directory "internal/handler" exists
      - directory "internal/service" exists
      - directory "internal/repository" exists

    medium:
      - directory "internal/model" exists
      - handlers call services, services call repos

  detection_commands:
    - "find . -type d -name 'handler'"
    - "find . -type d -name 'service'"
    - "find . -type d -name 'repository'"
```

---

### 2. Error Handling Patterns

#### pkg/errors Style

```yaml
pkg_errors:
  indicators:
    - import "github.com/pkg/errors"
    - usage of "errors.Wrap"
    - usage of "errors.WithStack"
    - usage of "errors.Cause"

  detection:
    grep_patterns:
      - '"github.com/pkg/errors"'
      - 'errors\.Wrap\('
      - 'errors\.WithStack\('

  example: |
    if err != nil {
        return errors.Wrap(err, "failed to get user")
    }
```

#### fmt.Errorf with %w

```yaml
fmt_errorf:
  indicators:
    - usage of "fmt.Errorf" with "%w"
    - error wrapping with context

  detection:
    grep_patterns:
      - 'fmt\.Errorf.*%w'
      - 'fmt\.Errorf\("[^"]*:.*%w'

  example: |
    if err != nil {
        return fmt.Errorf("get user by id %s: %w", id, err)
    }
```

#### Custom Error Types

```yaml
custom_errors:
  indicators:
    - "type.*Error struct" definitions
    - Error() method implementations
    - Sentinel errors (var Err* = errors.New)

  detection:
    grep_patterns:
      - 'type \w+Error struct'
      - 'var Err\w+ = errors\.New'
      - 'func \(.*Error\) Error\(\) string'

  example: |
    type NotFoundError struct {
        Resource string
        ID       string
    }

    func (e *NotFoundError) Error() string {
        return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
    }

    var ErrUserNotFound = errors.New("user not found")
```

---

### 3. Logging Patterns

#### slog (Go 1.21+)

```yaml
slog:
  detection:
    imports:
      - '"log/slog"'

    usage_patterns:
      - 'slog\.Info\('
      - 'slog\.Error\('
      - 'slog\.With\('
      - 'slog\.NewJSONHandler'

  example: |
    slog.Info("user created",
        slog.String("user_id", user.ID),
        slog.String("email", user.Email),
    )
```

#### zerolog

```yaml
zerolog:
  detection:
    imports:
      - '"github.com/rs/zerolog"'
      - '"github.com/rs/zerolog/log"'

    usage_patterns:
      - 'log\.Info\(\)'
      - 'log\.Error\(\)'
      - '\.Str\('
      - '\.Int\('
      - '\.Msg\('

  example: |
    log.Info().
        Str("user_id", user.ID).
        Str("email", user.Email).
        Msg("user created")
```

#### zap

```yaml
zap:
  detection:
    imports:
      - '"go.uber.org/zap"'

    usage_patterns:
      - 'zap\.NewProduction'
      - 'zap\.NewDevelopment'
      - 'logger\.Info\('
      - 'zap\.String\('
      - 'zap\.Int\('

  example: |
    logger.Info("user created",
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
    )
```

#### logrus

```yaml
logrus:
  detection:
    imports:
      - '"github.com/sirupsen/logrus"'

    usage_patterns:
      - 'logrus\.Info'
      - 'logrus\.WithFields'
      - 'log\.WithFields'
      - 'logrus\.Fields{'

  example: |
    logrus.WithFields(logrus.Fields{
        "user_id": user.ID,
        "email":   user.Email,
    }).Info("user created")
```

---

### 4. Database Patterns

#### GORM

```yaml
gorm:
  detection:
    imports:
      - '"gorm.io/gorm"'
      - '"gorm.io/driver/'

    usage_patterns:
      - 'gorm\.Open\('
      - '\.Create\(&'
      - '\.First\(&'
      - '\.Find\(&'
      - '\.Where\('
      - 'gorm\.Model'

    model_patterns:
      - 'gorm\.Model embedded'
      - 'gorm:"' tags

  example: |
    type User struct {
        gorm.Model
        Email    string `gorm:"uniqueIndex"`
        Name     string
    }

    db.Create(&user)
    db.First(&user, "id = ?", id)
```

#### sqlx

```yaml
sqlx:
  detection:
    imports:
      - '"github.com/jmoiron/sqlx"'

    usage_patterns:
      - 'sqlx\.Connect'
      - 'sqlx\.Open'
      - '\.Get\(&'
      - '\.Select\(&'
      - '\.NamedExec\('
      - 'db:"' tags

  example: |
    type User struct {
        ID    string `db:"id"`
        Email string `db:"email"`
    }

    db.Get(&user, "SELECT * FROM users WHERE id=$1", id)
```

#### pgx

```yaml
pgx:
  detection:
    imports:
      - '"github.com/jackc/pgx/v5"'
      - '"github.com/jackc/pgx/v5/pgxpool"'

    usage_patterns:
      - 'pgx\.Connect'
      - 'pgxpool\.New'
      - '\.QueryRow\('
      - '\.Scan\('

  example: |
    row := db.QueryRow(ctx, "SELECT id, email FROM users WHERE id=$1", id)
    err := row.Scan(&user.ID, &user.Email)
```

---

### 5. HTTP Framework Patterns

#### chi

```yaml
chi:
  detection:
    imports:
      - '"github.com/go-chi/chi"'
      - '"github.com/go-chi/chi/v5"'

    usage_patterns:
      - 'chi\.NewRouter\(\)'
      - 'r\.Get\('
      - 'r\.Post\('
      - 'r\.Route\('
      - 'chi\.URLParam\('

  example: |
    r := chi.NewRouter()
    r.Get("/users/{id}", handler.GetUser)
    r.Post("/users", handler.CreateUser)
```

#### gin

```yaml
gin:
  detection:
    imports:
      - '"github.com/gin-gonic/gin"'

    usage_patterns:
      - 'gin\.Default\(\)'
      - 'gin\.New\(\)'
      - 'c\.JSON\('
      - 'c\.Param\('
      - 'c\.Bind\('

  example: |
    r := gin.Default()
    r.GET("/users/:id", handler.GetUser)
    r.POST("/users", handler.CreateUser)
```

#### echo

```yaml
echo:
  detection:
    imports:
      - '"github.com/labstack/echo/v4"'

    usage_patterns:
      - 'echo\.New\(\)'
      - 'c\.JSON\('
      - 'c\.Param\('
      - 'c\.Bind\('

  example: |
    e := echo.New()
    e.GET("/users/:id", handler.GetUser)
    e.POST("/users", handler.CreateUser)
```

---

### 6. Configuration Patterns

#### viper

```yaml
viper:
  detection:
    imports:
      - '"github.com/spf13/viper"'

    usage_patterns:
      - 'viper\.SetConfigName'
      - 'viper\.ReadInConfig'
      - 'viper\.GetString'
      - 'viper\.Unmarshal'

  example: |
    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    viper.ReadInConfig()

    dbHost := viper.GetString("database.host")
```

#### envconfig

```yaml
envconfig:
  detection:
    imports:
      - '"github.com/kelseyhightower/envconfig"'

    usage_patterns:
      - 'envconfig\.Process'
      - 'envvar:"' tags

  example: |
    type Config struct {
        Port     int    `envconfig:"PORT" default:"8080"`
        DBHost   string `envconfig:"DB_HOST" required:"true"`
    }

    envconfig.Process("APP", &config)
```

---

### 7. Testing Patterns

#### Table-Driven Tests

```yaml
table_driven:
  detection:
    patterns:
      - 'tests := \[\]struct'
      - 'tt := range tests'
      - 't\.Run\(tt\.name'
      - 'tc := range testCases'

  example: |
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid", "input", "output", false},
        {"invalid", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
```

#### testify

```yaml
testify:
  detection:
    imports:
      - '"github.com/stretchr/testify/assert"'
      - '"github.com/stretchr/testify/require"'
      - '"github.com/stretchr/testify/mock"'
      - '"github.com/stretchr/testify/suite"'

    usage_patterns:
      - 'assert\.Equal'
      - 'assert\.NoError'
      - 'require\.NoError'
      - 'mock\.Mock'

  example: |
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    require.NotNil(t, result)
```

---

## Detection Algorithm

```python
def detect_patterns(codebase_path):
    results = {}

    # 1. Scan go.mod for dependencies
    deps = parse_go_mod(codebase_path)

    # 2. Scan imports across all files
    imports = scan_all_imports(codebase_path)

    # 3. For each category, run detection
    for category in [architecture, error_handling, logging, database, http, config, testing]:
        for pattern in category.patterns:
            confidence = 0.0

            # Check indicators
            for indicator in pattern.strong_indicators:
                if check_indicator(indicator, codebase_path, imports, deps):
                    confidence += 0.3

            for indicator in pattern.medium_indicators:
                if check_indicator(indicator, codebase_path, imports, deps):
                    confidence += 0.2

            for indicator in pattern.weak_indicators:
                if check_indicator(indicator, codebase_path, imports, deps):
                    confidence += 0.1

            # Record if above threshold
            if confidence >= pattern.confidence_threshold:
                results[category.name] = {
                    "pattern": pattern.name,
                    "confidence": confidence,
                    "evidence": collected_evidence
                }

    return results
```

---

## Output Format

```json
{
  "detected_patterns": {
    "architecture": {
      "pattern": "clean_architecture",
      "confidence": 0.8,
      "evidence": [
        "internal/domain/ exists",
        "internal/usecase/ exists",
        "interfaces defined in domain"
      ]
    },
    "error_handling": {
      "pattern": "fmt_errorf",
      "confidence": 0.9,
      "evidence": [
        "45 occurrences of fmt.Errorf with %w"
      ]
    },
    "logging": {
      "pattern": "zerolog",
      "confidence": 1.0,
      "evidence": [
        "import github.com/rs/zerolog in 23 files"
      ]
    },
    "database": {
      "pattern": "gorm",
      "confidence": 1.0,
      "evidence": [
        "import gorm.io/gorm",
        "gorm.Model used in 5 structs"
      ]
    },
    "http_framework": {
      "pattern": "chi",
      "confidence": 1.0,
      "evidence": [
        "import github.com/go-chi/chi/v5"
      ]
    },
    "testing": {
      "pattern": "table_driven_testify",
      "confidence": 0.85,
      "evidence": [
        "table-driven pattern in 18 test files",
        "testify/assert used"
      ]
    }
  }
}
```
