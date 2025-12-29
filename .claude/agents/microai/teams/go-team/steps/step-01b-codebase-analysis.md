---
stepNumber: 1.5
nextStep: './step-02-requirements.md'
agent: orchestrator
hasBreakpoint: false
conditional: true
condition: "existing_codebase_detected"
checkpoint:
  enabled: true
  id_format: "cp-01b-analysis"
---

# Step 01b: Codebase Analysis

## STEP GOAL

Analyze existing codebase để extract patterns, conventions, và context. Step này chỉ chạy khi detect được existing Go code.

## TRIGGER CONDITION

```yaml
trigger:
  - go.mod exists
  - AND (internal/ OR pkg/) contains .go files
  - AND total .go files > 5
```

## EXECUTION SEQUENCE

### 1. Detection

```
═══════════════════════════════════════════════════════════
🔍 EXISTING CODEBASE DETECTED
═══════════════════════════════════════════════════════════

Go module: {module_name}
Go version: {go_version}
Go files: {count} files
Lines of code: ~{loc}

Mode: EXTEND (adapting to existing codebase)

Running analysis...

═══════════════════════════════════════════════════════════
```

### 2. Structure Analysis

```bash
# Commands to run
tree -L 3 -d  # Directory structure
find . -name "*.go" | wc -l  # File count
wc -l $(find . -name "*.go") | tail -1  # Lines of code
```

**Output:**
```
📂 STRUCTURE ANALYSIS
─────────────────────────────────────────────────────────

Directory Layout:
├── cmd/
│   └── api/main.go          (entry point)
├── internal/
│   ├── handler/             (12 files)
│   ├── service/             (8 files)
│   ├── repository/          (6 files)
│   └── model/               (5 files)
├── pkg/
│   └── middleware/          (3 files)
├── configs/
└── tests/

Structure Type: Standard Go Layout
Entry Points: 1 (cmd/api/main.go)

─────────────────────────────────────────────────────────
```

### 3. Pattern Detection

Load pattern detector từ `../codebase/pattern-detector.md`

```bash
# Scan imports
grep -rh "^import" --include="*.go" . | sort | uniq -c | sort -rn

# Detect patterns
grep -r "github.com/rs/zerolog" --include="*.go" .
grep -r "gorm.io/gorm" --include="*.go" .
grep -r "github.com/go-chi/chi" --include="*.go" .
```

**Output:**
```
🔧 PATTERN DETECTION
─────────────────────────────────────────────────────────

Architecture:     Clean Architecture (confidence: 85%)
                  └─ Evidence: domain/, usecase/, repository/ structure

Error Handling:   fmt.Errorf with %w (confidence: 92%)
                  └─ 67 occurrences across codebase

Logging:          zerolog (confidence: 100%)
                  └─ Used in 28 files

Database:         GORM (confidence: 100%)
                  └─ 15 models with gorm.Model

HTTP Framework:   chi (confidence: 100%)
                  └─ Router in cmd/api/main.go

Config:           viper (confidence: 100%)
                  └─ config.yaml + environment

Testing:          Table-driven + testify (confidence: 88%)
                  └─ 45 test files

─────────────────────────────────────────────────────────
```

### 4. Interface Extraction

```bash
# Find all interfaces
grep -rn "type.*interface {" --include="*.go" internal/
```

**Output:**
```
📋 EXISTING INTERFACES
─────────────────────────────────────────────────────────

Repository Layer:
┌─────────────────────────────────────────────────────────┐
│ UserRepository (internal/repository/user.go)            │
├─────────────────────────────────────────────────────────┤
│ • Create(ctx, user *User) error                         │
│ • GetByID(ctx, id string) (*User, error)                │
│ • GetByEmail(ctx, email string) (*User, error)          │
│ • Update(ctx, user *User) error                         │
│ • Delete(ctx, id string) error                          │
│ • List(ctx, opts ListOptions) ([]*User, error)          │
└─────────────────────────────────────────────────────────┘

Service Layer:
┌─────────────────────────────────────────────────────────┐
│ AuthService (internal/service/auth.go)                  │
├─────────────────────────────────────────────────────────┤
│ • Login(ctx, email, password string) (*Token, error)    │
│ • Logout(ctx, token string) error                       │
│ • ValidateToken(ctx, token string) (*Claims, error)     │
│ • RefreshToken(ctx, token string) (*Token, error)       │
└─────────────────────────────────────────────────────────┘

Existing Types:
• User (ID, Email, PasswordHash, Name, CreatedAt, UpdatedAt)
• Token (AccessToken, RefreshToken, ExpiresAt)
• Claims (UserID, Email, Role, ExpiresAt)

─────────────────────────────────────────────────────────
```

### 5. Style Extraction

Load style extractor từ `../codebase/style-extractor.md`

**Output:**
```
🎨 CODING STYLE
─────────────────────────────────────────────────────────

File Naming:      snake_case
                  └─ user_repository.go, auth_handler.go

Imports:          Grouped (stdlib → external → internal)
                  └─ Blank line separators

Functions:
• Context:        First parameter, named "ctx"
• Errors:         Last return value
• Receivers:      Pointer (*Type)

Error Handling:
  if err != nil {
      return fmt.Errorf("operation: %w", err)
  }

Logging (zerolog):
  log.Info().
      Str("user_id", id).
      Msg("action completed")

Struct Tags:      json:"snake_case" db:"snake_case"

Test Style:       Table-driven with testify
                  Naming: Test{Function}_{Scenario}

─────────────────────────────────────────────────────────
```

### 6. Dependency Mapping

```bash
# Parse go.mod
cat go.mod

# Key dependencies
go list -m all | head -20
```

**Output:**
```
📦 DEPENDENCIES
─────────────────────────────────────────────────────────

Module: github.com/myorg/myproject
Go Version: 1.21

Key Dependencies:
├── HTTP:     github.com/go-chi/chi/v5 v5.0.10
├── Logging:  github.com/rs/zerolog v1.31.0
├── Database: gorm.io/gorm v1.25.5
│             gorm.io/driver/postgres v1.5.4
├── Config:   github.com/spf13/viper v1.17.0
├── Validate: github.com/go-playground/validator/v10 v10.16.0
└── Testing:  github.com/stretchr/testify v1.8.4

─────────────────────────────────────────────────────────
```

### 7. Generate Context for Agents

```yaml
agent_context:
  pm_agent:
    existing_features:
      - "User management (CRUD)"
      - "Authentication (login/logout/token)"
      - "Basic authorization"
    potential_extensions:
      - "Password reset"
      - "Email verification"
      - "OAuth integration"

  architect_agent:
    current_architecture: "clean_architecture"
    existing_layers:
      handler: "internal/handler/"
      service: "internal/service/"
      repository: "internal/repository/"
      model: "internal/model/"
    interfaces_to_extend:
      - "AuthService: add password reset methods"
      - "UserRepository: no changes needed"
    patterns_to_follow:
      - "Dependency injection via constructors"
      - "Return (*Type, error) for queries"

  coder_agent:
    style_guide: "{extracted_style}"
    imports_template: "{extracted_import_order}"
    error_template: "fmt.Errorf(\"{operation}: %w\", err)"
    types_to_reuse: ["User", "Token", "Claims"]
    avoid:
      - "Creating duplicate User type"
      - "Different error handling style"

  test_agent:
    test_pattern: "table_driven"
    framework: "testify"
    existing_mocks: ["MockUserRepository", "MockAuthService"]
    example_test: "{extracted_test_example}"

  reviewer_agent:
    style_checklist:
      - "snake_case file names"
      - "Grouped imports"
      - "Error wrapping with %w"
      - "zerolog structured logging"
```

### 8. Analysis Summary

```
═══════════════════════════════════════════════════════════
✓ CODEBASE ANALYSIS COMPLETE
═══════════════════════════════════════════════════════════

Project: github.com/myorg/myproject
Mode: EXTEND

Analysis Summary:
┌─────────────────────────────────────────────────────────┐
│ Files:          89 Go files, ~12,450 LOC                │
│ Architecture:   Clean Architecture                      │
│ Error Handling: fmt.Errorf with %w                      │
│ Logging:        zerolog                                 │
│ Database:       GORM + PostgreSQL                       │
│ HTTP:           chi router                              │
│ Testing:        Table-driven + testify                  │
├─────────────────────────────────────────────────────────┤
│ Interfaces:     8 extracted                             │
│ Types:          12 models identified                    │
│ Style Rules:    15 conventions detected                 │
└─────────────────────────────────────────────────────────┘

Context Injection:
✓ PM Agent: Feature gaps identified
✓ Architect Agent: Extension points mapped
✓ Coder Agent: Style guide generated
✓ Test Agent: Test patterns extracted
✓ Reviewer Agent: Checklist prepared

Recommendations:
• Extend AuthService for new auth features
• Reuse existing User and Token models
• Follow established error handling pattern
• Maintain zerolog structured logging

═══════════════════════════════════════════════════════════

Press [Enter] to continue to Requirements phase...
```

## OUTPUT

```yaml
outputs:
  analysis:
    structure:
      type: "standard_layout"
      entry_points: [...]
      packages: {...}

    patterns:
      architecture: "clean_architecture"
      error_handling: "fmt_errorf"
      logging: "zerolog"
      database: "gorm"
      http: "chi"
      testing: "table_driven_testify"

    interfaces:
      - name: "UserRepository"
        methods: [...]
      - name: "AuthService"
        methods: [...]

    types:
      - name: "User"
        fields: [...]

    style_guide:
      file_naming: "snake_case"
      imports: "grouped"
      error_format: "fmt.Errorf with %w"
      ...

    agent_context:
      pm: {...}
      architect: {...}
      coder: {...}
      test: {...}
      reviewer: {...}

    recommendations:
      extend: ["AuthService"]
      reuse: ["User", "Token"]
      avoid: ["duplicate types"]
```

## SUCCESS CRITERIA

- [ ] Structure analyzed
- [ ] Patterns detected
- [ ] Interfaces extracted
- [ ] Style guide generated
- [ ] Agent contexts prepared
- [ ] Recommendations ready

---

## SKIP CONDITION

Skip this step if:
- No go.mod found
- No .go files in internal/ or pkg/
- Fewer than 5 Go files (likely new project)

When skipped:
```
Mode: GREENFIELD (new project)
Skipping codebase analysis...
Proceeding to Requirements phase...
```

---

## NEXT STEP

Load `./step-02-requirements.md` with injected codebase context
