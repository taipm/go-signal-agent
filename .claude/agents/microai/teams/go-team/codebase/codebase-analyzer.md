# Codebase Analyzer

**Version:** 1.0.0

Phân tích codebase Go hiện có để extract patterns, conventions, và context cho các agents.

---

## Quick Start

```bash
# Run full analysis
*analyze

# Specific analysis
*analyze:structure    # Directory structure
*analyze:patterns     # Code patterns
*analyze:interfaces   # Existing interfaces
*analyze:deps         # Dependencies
*analyze:style        # Coding style
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    CODEBASE ANALYZER                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Structure       │  │ Pattern         │  │ Style       │ │
│  │ Scanner         │  │ Detector        │  │ Extractor   │ │
│  │                 │  │                 │  │             │ │
│  │ • Directories   │  │ • Architecture  │  │ • Naming    │ │
│  │ • Files         │  │ • Error handling│  │ • Imports   │ │
│  │ • Packages      │  │ • Logging       │  │ • Comments  │ │
│  │ • Entry points  │  │ • Database      │  │ • Formatting│ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Interface       │  │ Dependency      │  │ Context     │ │
│  │ Extractor       │  │ Mapper          │  │ Builder     │ │
│  │                 │  │                 │  │             │ │
│  │ • Types         │  │ • go.mod        │  │ • For PM    │ │
│  │ • Methods       │  │ • Imports       │  │ • For Arch  │ │
│  │ • Contracts     │  │ • Graph         │  │ • For Coder │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Analysis Components

### 1. Structure Scanner

Scans directory layout và file organization.

```yaml
structure_scan:
  inputs:
    - Root directory path
    - Exclude patterns (.git, vendor, node_modules)

  outputs:
    directory_tree:
      cmd/: "Entry points"
      internal/: "Private packages"
      pkg/: "Public packages"
      configs/: "Configuration"
      docs/: "Documentation"
      tests/: "Test files"

    file_counts:
      go_files: 89
      test_files: 34
      config_files: 5

    entry_points:
      - cmd/api/main.go
      - cmd/worker/main.go
```

### 2. Pattern Detector

Identifies architectural patterns và conventions.

```yaml
pattern_detection:
  architecture:
    patterns:
      - clean_architecture
      - hexagonal
      - mvc
      - simple_layered
      - microservices

    detection_rules:
      clean_architecture:
        indicators:
          - "internal/domain/" exists
          - "internal/usecase/" exists
          - "internal/repository/" exists
          - Interfaces defined in domain

      hexagonal:
        indicators:
          - "internal/ports/" exists
          - "internal/adapters/" exists
          - Clear port/adapter separation

  error_handling:
    patterns:
      - pkg_errors      # github.com/pkg/errors
      - fmt_errorf      # fmt.Errorf with %w
      - custom_errors   # Custom error types
      - sentinel_errors # var ErrNotFound = errors.New(...)

    detection:
      scan_for:
        - "errors.Wrap"
        - "errors.WithStack"
        - "fmt.Errorf.*%w"
        - "type.*Error struct"

  logging:
    libraries:
      - slog      # Go 1.21+ structured logging
      - zerolog   # github.com/rs/zerolog
      - zap       # go.uber.org/zap
      - logrus    # github.com/sirupsen/logrus
      - log       # Standard library

    detection:
      scan_imports_for:
        - "log/slog"
        - "github.com/rs/zerolog"
        - "go.uber.org/zap"
        - "github.com/sirupsen/logrus"

  database:
    libraries:
      - gorm      # gorm.io/gorm
      - sqlx      # github.com/jmoiron/sqlx
      - pgx       # github.com/jackc/pgx
      - ent       # entgo.io/ent
      - raw_sql   # database/sql only

    detection:
      scan_imports_for:
        - "gorm.io/gorm"
        - "github.com/jmoiron/sqlx"
        - "github.com/jackc/pgx"
        - "entgo.io/ent"

  http_framework:
    libraries:
      - chi         # github.com/go-chi/chi
      - gin         # github.com/gin-gonic/gin
      - echo        # github.com/labstack/echo
      - fiber       # github.com/gofiber/fiber
      - gorilla_mux # github.com/gorilla/mux
      - stdlib      # net/http only

  config:
    libraries:
      - viper   # github.com/spf13/viper
      - envconfig # github.com/kelseyhightower/envconfig
      - godotenv # github.com/joho/godotenv
      - yaml    # gopkg.in/yaml.v3
```

### 3. Interface Extractor

Extracts existing interfaces và types.

```yaml
interface_extraction:
  scan:
    - All .go files
    - Focus on internal/ and pkg/

  extract:
    interfaces:
      - name: "UserRepository"
        package: "internal/repository"
        methods:
          - "Create(ctx context.Context, user *User) error"
          - "GetByID(ctx context.Context, id string) (*User, error)"
          - "GetByEmail(ctx context.Context, email string) (*User, error)"
          - "Update(ctx context.Context, user *User) error"
          - "Delete(ctx context.Context, id string) error"

      - name: "AuthService"
        package: "internal/service"
        methods:
          - "Login(ctx context.Context, email, password string) (*Token, error)"
          - "Logout(ctx context.Context, token string) error"
          - "ValidateToken(ctx context.Context, token string) (*Claims, error)"

    structs:
      - name: "User"
        package: "internal/model"
        fields:
          - "ID string"
          - "Email string"
          - "PasswordHash string"
          - "CreatedAt time.Time"
          - "UpdatedAt time.Time"

    type_aliases:
      - "type UserID = string"
      - "type Email = string"
```

### 4. Style Extractor

Extracts coding conventions và style.

```yaml
style_extraction:
  naming:
    files:
      pattern: "snake_case"  # user_repository.go
      examples:
        - "user_repository.go"
        - "auth_handler.go"

    packages:
      pattern: "lowercase"   # repository, service
      examples:
        - "repository"
        - "service"

    types:
      pattern: "PascalCase"  # UserRepository
      examples:
        - "UserRepository"
        - "AuthService"

    functions:
      exported: "PascalCase"
      private: "camelCase"

    variables:
      exported: "PascalCase"
      private: "camelCase"
      constants: "PascalCase or ALL_CAPS"

  imports:
    order:
      - "Standard library"
      - "External packages"
      - "Internal packages"
    grouping: "separated by blank line"
    example: |
      import (
          "context"
          "fmt"

          "github.com/rs/zerolog"

          "myproject/internal/model"
      )

  comments:
    exported_functions: "godoc style"
    inline: "minimal"
    example: |
      // GetByID retrieves a user by their unique identifier.
      // Returns ErrNotFound if user does not exist.
      func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {

  error_messages:
    format: "lowercase, no punctuation"
    example: "failed to get user"

  context_usage:
    first_param: true
    naming: "ctx"
```

### 5. Dependency Mapper

Maps dependencies và import graph.

```yaml
dependency_mapping:
  go_mod:
    module: "github.com/myorg/myproject"
    go_version: "1.21"
    dependencies:
      direct:
        - "github.com/go-chi/chi/v5 v5.0.10"
        - "github.com/rs/zerolog v1.31.0"
        - "gorm.io/gorm v1.25.5"
      indirect:
        - "..."

  import_graph:
    cmd/api/main.go:
      - internal/handler
      - internal/service
      - internal/repository
      - pkg/middleware

    internal/handler/user.go:
      - internal/service
      - internal/model

    internal/service/user.go:
      - internal/repository
      - internal/model

  package_dependencies:
    handler:
      depends_on: [service, model]
      depended_by: [main]

    service:
      depends_on: [repository, model]
      depended_by: [handler]

    repository:
      depends_on: [model]
      depended_by: [service]
```

---

## Analysis Output

### Complete Analysis Report

```json
{
  "project": {
    "name": "my-api-service",
    "module": "github.com/myorg/myproject",
    "go_version": "1.21",
    "analyzed_at": "2025-12-28T23:00:00Z"
  },

  "metrics": {
    "total_files": 89,
    "go_files": 67,
    "test_files": 22,
    "lines_of_code": 12450,
    "packages": 15
  },

  "structure": {
    "type": "standard",
    "entry_points": ["cmd/api/main.go"],
    "directories": {
      "cmd": "entry points",
      "internal": "private code",
      "pkg": "public libraries",
      "configs": "configuration"
    }
  },

  "patterns": {
    "architecture": "clean_architecture",
    "error_handling": "fmt_errorf",
    "logging": "zerolog",
    "database": "gorm",
    "http_framework": "chi",
    "config": "viper"
  },

  "interfaces": [
    {
      "name": "UserRepository",
      "package": "internal/repository",
      "methods": ["Create", "GetByID", "GetByEmail", "Update", "Delete"]
    }
  ],

  "types": [
    {
      "name": "User",
      "package": "internal/model",
      "fields": ["ID", "Email", "PasswordHash", "CreatedAt", "UpdatedAt"]
    }
  ],

  "style": {
    "file_naming": "snake_case",
    "import_order": ["stdlib", "external", "internal"],
    "error_format": "lowercase_no_punctuation",
    "context_first_param": true
  },

  "recommendations": {
    "extend_interfaces": ["UserRepository", "AuthService"],
    "reuse_types": ["User", "Token"],
    "follow_patterns": ["error handling with %w", "zerolog structured logging"],
    "avoid": ["creating duplicate functionality"]
  }
}
```

---

## Context Injection

### For PM Agent

```yaml
pm_context:
  existing_features:
    - "User authentication (login/logout)"
    - "User CRUD operations"
    - "Token-based auth"

  gaps_identified:
    - "No password reset"
    - "No email verification"
    - "No rate limiting"

  constraints:
    - "Must integrate with existing UserService"
    - "Must use existing User model"
```

### For Architect Agent

```yaml
architect_context:
  current_architecture: "clean_architecture"

  existing_layers:
    handlers:
      path: "internal/handler"
      existing: ["user_handler.go", "auth_handler.go"]
    services:
      path: "internal/service"
      existing: ["user_service.go", "auth_service.go"]
    repositories:
      path: "internal/repository"
      existing: ["user_repository.go"]
    models:
      path: "internal/model"
      existing: ["user.go", "token.go"]

  interfaces_to_extend:
    - name: "AuthService"
      add_methods:
        - "ResetPassword(ctx, email string) error"
        - "VerifyResetToken(ctx, token string) (*User, error)"

  patterns_to_follow:
    - "Dependency injection via constructors"
    - "Interfaces in same package as implementation"
    - "Return (*Type, error) for queries"
```

### For Coder Agent

```yaml
coder_context:
  style_guide:
    file_naming: "snake_case"
    error_handling: |
      if err != nil {
          return fmt.Errorf("operation description: %w", err)
      }
    logging: |
      log.Info().
          Str("user_id", userID).
          Msg("operation completed")
    context: "Always first parameter, named 'ctx'"

  imports_template: |
    import (
        "context"
        "fmt"

        "github.com/rs/zerolog/log"

        "github.com/myorg/myproject/internal/model"
        "github.com/myorg/myproject/internal/repository"
    )

  existing_code_examples:
    service_method: |
      func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
          user, err := s.repo.GetByID(ctx, id)
          if err != nil {
              return nil, fmt.Errorf("get user by id: %w", err)
          }
          return user, nil
      }

  types_to_reuse:
    - "model.User"
    - "model.Token"
    - "repository.UserRepository"

  avoid:
    - "Creating new User type"
    - "Different error handling style"
    - "Different logging format"
```

### For Test Agent

```yaml
test_context:
  test_patterns:
    style: "table_driven"
    framework: "testify"
    mocking: "testify/mock"

  existing_test_example: |
    func TestUserService_GetByID(t *testing.T) {
        tests := []struct {
            name    string
            id      string
            want    *model.User
            wantErr bool
        }{
            {
                name: "success",
                id:   "user-123",
                want: &model.User{ID: "user-123"},
            },
            {
                name:    "not found",
                id:      "nonexistent",
                wantErr: true,
            },
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                // Arrange
                mockRepo := new(MockUserRepository)
                svc := NewUserService(mockRepo)

                // Act
                got, err := svc.GetByID(context.Background(), tt.id)

                // Assert
                if tt.wantErr {
                    assert.Error(t, err)
                } else {
                    assert.NoError(t, err)
                    assert.Equal(t, tt.want, got)
                }
            })
        }
    }

  existing_mocks:
    - "MockUserRepository"
    - "MockAuthService"

  test_utilities:
    - "testutil.NewTestDB()"
    - "testutil.CreateTestUser()"
```

---

## Commands

### Analysis Commands

| Command | Description |
|---------|-------------|
| `*analyze` | Run full codebase analysis |
| `*analyze:structure` | Analyze directory structure |
| `*analyze:patterns` | Detect code patterns |
| `*analyze:interfaces` | List existing interfaces |
| `*analyze:types` | List existing types/models |
| `*analyze:deps` | Show dependencies |
| `*analyze:style` | Extract style conventions |
| `*analyze:report` | Generate full report |

### Context Commands

| Command | Description |
|---------|-------------|
| `*context:show` | Show injected context |
| `*context:refresh` | Re-analyze and refresh |
| `*context:export` | Export context to file |

### Integration Commands

| Command | Description |
|---------|-------------|
| `*extend:{interface}` | Show how to extend interface |
| `*reuse:{type}` | Show how to reuse type |
| `*follow:{pattern}` | Show pattern examples |

---

## Integration with Workflow

### Automatic Detection

```yaml
on_session_start:
  1. Check if go.mod exists
  2. Check if internal/ or pkg/ has Go files
  3. IF existing code found:
     - Set mode = "extend"
     - Trigger codebase analysis
     - Inject context to agents
  4. ELSE:
     - Set mode = "greenfield"
     - Skip analysis
```

### Mode Indicator

```
═══════════════════════════════════════════════════════════
🔍 EXISTING CODEBASE DETECTED
═══════════════════════════════════════════════════════════

Mode: EXTEND (not greenfield)

Project: github.com/myorg/myproject
Files: 89 Go files
Architecture: Clean Architecture

Running analysis...

═══════════════════════════════════════════════════════════
```

---

## Benefits

1. **Consistency:** New code matches existing patterns
2. **Reusability:** Leverages existing components
3. **Integration:** Seamless fit with current codebase
4. **Quality:** Follows established conventions
5. **Speed:** Less refactoring needed post-generation
