# Style Extractor

**Version:** 1.0.0

Extracts coding style và conventions từ existing codebase để đảm bảo consistency.

---

## Extraction Categories

### 1. File Naming Conventions

```yaml
file_naming:
  detection:
    scan: "all .go files in internal/ and pkg/"

    patterns:
      snake_case:
        regex: "^[a-z][a-z0-9]*(_[a-z0-9]+)*\\.go$"
        examples: ["user_repository.go", "auth_handler.go"]

      camelCase:
        regex: "^[a-z][a-zA-Z0-9]*\\.go$"
        examples: ["userRepository.go", "authHandler.go"]

      kebab_case:
        regex: "^[a-z][a-z0-9]*(-[a-z0-9]+)*\\.go$"
        examples: ["user-repository.go"]

  output:
    dominant_pattern: "snake_case"
    confidence: 0.95
    examples:
      - "user_repository.go"
      - "auth_service.go"
      - "token_handler.go"
```

### 2. Package Naming

```yaml
package_naming:
  detection:
    scan: "package declarations in all .go files"

    rules:
      - "lowercase only"
      - "no underscores (usually)"
      - "short, concise names"

  output:
    style: "lowercase_short"
    examples:
      - "package handler"
      - "package service"
      - "package repository"
      - "package model"
```

### 3. Import Organization

```yaml
import_organization:
  detection:
    scan: "import blocks in all .go files"

  patterns:
    grouped:
      description: "Imports grouped by category with blank lines"
      order: ["stdlib", "external", "internal"]
      example: |
        import (
            "context"
            "fmt"
            "time"

            "github.com/rs/zerolog"
            "gorm.io/gorm"

            "myproject/internal/model"
            "myproject/internal/repository"
        )

    ungrouped:
      description: "All imports together"
      example: |
        import (
            "context"
            "github.com/rs/zerolog"
            "myproject/internal/model"
        )

  output:
    style: "grouped"
    order: ["stdlib", "external", "internal"]
    separator: "blank_line"
```

### 4. Function Signatures

```yaml
function_signatures:
  detection:
    scan: "all function declarations"

  patterns:
    context_first:
      description: "context.Context as first parameter"
      regex: "func.*\\(ctx context\\.Context"
      prevalence: 0.95

    error_last:
      description: "error as last return value"
      regex: "\\) \\(.*error\\)$|\\) error$"
      prevalence: 0.98

    pointer_receivers:
      description: "Methods use pointer receivers"
      regex: "func \\(\\w+ \\*\\w+\\)"
      prevalence: 0.90

  output:
    context_position: "first"
    context_name: "ctx"
    error_position: "last"
    receiver_style: "pointer"

    template: |
      func (s *ServiceName) MethodName(ctx context.Context, param Type) (*Result, error) {
```

### 5. Error Handling Style

```yaml
error_handling:
  detection:
    scan: "error handling blocks"

  patterns:
    early_return:
      description: "Return immediately on error"
      example: |
        if err != nil {
            return nil, fmt.Errorf("operation: %w", err)
        }

    error_wrapping:
      styles:
        fmt_errorf:
          pattern: 'fmt.Errorf("description: %w", err)'
          message_style: "lowercase, no period"

        pkg_errors:
          pattern: 'errors.Wrap(err, "description")'
          message_style: "lowercase, no period"

    error_message_format:
      examples:
        - "get user by id: %w"
        - "create user: %w"
        - "validate input: %w"

  output:
    style: "early_return"
    wrapping: "fmt_errorf"
    message_format: "operation description: %w"
    message_case: "lowercase"
    message_punctuation: "none"

    template: |
      if err != nil {
          return nil, fmt.Errorf("{operation}: %w", err)
      }
```

### 6. Logging Style

```yaml
logging_style:
  detection:
    scan: "logging statements"

  patterns:
    zerolog:
      chained_style: |
        log.Info().
            Str("user_id", userID).
            Str("action", "created").
            Msg("user operation completed")

    slog:
      args_style: |
        slog.Info("user operation completed",
            slog.String("user_id", userID),
            slog.String("action", "created"),
        )

  output:
    library: "zerolog"
    style: "chained"
    field_order: ["identifiers", "context", "metrics"]
    message_style: "descriptive action"

    template: |
      log.{Level}().
          Str("{id_field}", {id_value}).
          Str("{context_field}", {context_value}).
          Msg("{action description}")
```

### 7. Struct Definitions

```yaml
struct_definitions:
  detection:
    scan: "type struct definitions"

  patterns:
    field_ordering:
      order: ["id", "core_fields", "timestamps", "relations"]

    tag_style:
      json: "snake_case"
      db: "snake_case"
      gorm: "present for ORM"

    embedding:
      common: ["gorm.Model", "BaseModel"]

  output:
    field_order:
      - "ID (first)"
      - "Core business fields"
      - "Timestamps (CreatedAt, UpdatedAt)"
      - "Soft delete (DeletedAt)"
      - "Relations (last)"

    tag_format: 'json:"{snake_case}" db:"{snake_case}"'

    template: |
      type EntityName struct {
          ID        string    `json:"id" db:"id"`
          Name      string    `json:"name" db:"name"`
          CreatedAt time.Time `json:"created_at" db:"created_at"`
          UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
      }
```

### 8. Interface Definitions

```yaml
interface_definitions:
  detection:
    scan: "type interface definitions"

  patterns:
    location:
      - "Same package as implementation"
      - "Separate interfaces package"
      - "Domain/ports package"

    naming:
      suffix: "er" for single method (Reader, Writer)
      prefix: none usually
      descriptive: for multi-method (UserRepository)

    method_style:
      context: "first param"
      returns: "(Type, error) or error"

  output:
    location: "same_package"
    naming_convention: "descriptive"

    template: |
      type EntityRepository interface {
          Create(ctx context.Context, entity *Entity) error
          GetByID(ctx context.Context, id string) (*Entity, error)
          Update(ctx context.Context, entity *Entity) error
          Delete(ctx context.Context, id string) error
      }
```

### 9. Comment Style

```yaml
comment_style:
  detection:
    scan: "comments in code"

  patterns:
    exported:
      style: "godoc"
      format: "// FunctionName description..."

    inline:
      style: "minimal"
      usage: "only for complex logic"

    todos:
      format: "// TODO: description"

  output:
    exported_format: |
      // MethodName performs the described action.
      // It returns error if the operation fails.
      func (s *Service) MethodName(ctx context.Context) error {

    inline_format: "// explain why, not what"

    template: |
      // {FunctionName} {brief description}.
      // {Additional details if needed}.
```

### 10. Testing Style

```yaml
testing_style:
  detection:
    scan: "*_test.go files"

  patterns:
    test_naming:
      format: "Test{Function}_{Scenario}"
      examples:
        - "TestCreateUser_Success"
        - "TestCreateUser_DuplicateEmail"
        - "TestGetByID_NotFound"

    structure:
      pattern: "arrange_act_assert"
      sections:
        - "// Arrange"
        - "// Act"
        - "// Assert"

    table_driven:
      struct_name: "tests" or "testCases"
      fields: ["name", "input(s)", "want", "wantErr"]
      iterator: "tt" or "tc"

  output:
    naming: "Test{Function}_{Scenario}"
    structure: "AAA"
    style: "table_driven"

    template: |
      func TestServiceName_MethodName(t *testing.T) {
          tests := []struct {
              name    string
              input   InputType
              want    WantType
              wantErr bool
          }{
              {
                  name:  "success case",
                  input: validInput,
                  want:  expectedOutput,
              },
              {
                  name:    "error case",
                  input:   invalidInput,
                  wantErr: true,
              },
          }

          for _, tt := range tests {
              t.Run(tt.name, func(t *testing.T) {
                  // Arrange
                  svc := NewService(mockDeps)

                  // Act
                  got, err := svc.Method(context.Background(), tt.input)

                  // Assert
                  if tt.wantErr {
                      assert.Error(t, err)
                      return
                  }
                  assert.NoError(t, err)
                  assert.Equal(t, tt.want, got)
              })
          }
      }
```

---

## Extraction Process

```python
def extract_style(codebase_path):
    style_guide = {}

    # 1. Collect all Go files
    go_files = find_all_go_files(codebase_path)

    # 2. Extract each category
    style_guide["file_naming"] = analyze_file_names(go_files)
    style_guide["package_naming"] = analyze_packages(go_files)
    style_guide["imports"] = analyze_imports(go_files)
    style_guide["functions"] = analyze_functions(go_files)
    style_guide["errors"] = analyze_error_handling(go_files)
    style_guide["logging"] = analyze_logging(go_files)
    style_guide["structs"] = analyze_structs(go_files)
    style_guide["interfaces"] = analyze_interfaces(go_files)
    style_guide["comments"] = analyze_comments(go_files)
    style_guide["testing"] = analyze_tests(go_files)

    # 3. Generate templates
    style_guide["templates"] = generate_templates(style_guide)

    return style_guide
```

---

## Output: Generated Style Guide

```markdown
# Extracted Style Guide

## File Naming
- Pattern: snake_case
- Examples: user_repository.go, auth_handler.go

## Imports
Order: stdlib → external → internal
Separator: blank line between groups

## Functions
- Context: first parameter, named "ctx"
- Errors: last return value
- Receivers: pointer (*Type)

## Error Handling
```go
if err != nil {
    return fmt.Errorf("operation description: %w", err)
}
```

## Logging (zerolog)
```go
log.Info().
    Str("field", value).
    Msg("action completed")
```

## Structs
```go
type Entity struct {
    ID        string    `json:"id"`
    Field     Type      `json:"field"`
    CreatedAt time.Time `json:"created_at"`
}
```

## Tests
```go
func TestFunction_Scenario(t *testing.T) {
    tests := []struct{ name string; ... }{...}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {...})
    }
}
```
```

---

## Integration

Style guide is injected to:
- **Coder Agent**: For code generation
- **Test Agent**: For test generation
- **Reviewer Agent**: For style checking
