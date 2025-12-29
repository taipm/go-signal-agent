# Testing Patterns - Shared Knowledge

**Version:** 1.0.0
**Applies to:** All Agents

---

## TL;DR

- Table-driven tests cho multiple cases
- Naming: `Test{Function}_{Scenario}`
- AAA pattern: Arrange, Act, Assert
- Mock interfaces với testify/mock
- `t.Parallel()` cho independent tests

---

## 1. Table-Driven Tests

### Basic Structure

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed numbers", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### With Error Cases

```go
func TestParseUser(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *User
        wantErr bool
    }{
        {
            name:  "valid user",
            input: `{"name": "John", "email": "john@example.com"}`,
            want:  &User{Name: "John", Email: "john@example.com"},
        },
        {
            name:    "invalid json",
            input:   `{invalid}`,
            wantErr: true,
        },
        {
            name:    "missing required field",
            input:   `{"name": "John"}`,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseUser(tt.input)

            if tt.wantErr {
                if err == nil {
                    t.Errorf("ParseUser() expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Errorf("ParseUser() unexpected error: %v", err)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseUser() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## 2. Test Naming Convention

### Function Tests

```go
// Pattern: Test{Function}_{Scenario}

func TestCreateUser_Success(t *testing.T) { ... }
func TestCreateUser_DuplicateEmail(t *testing.T) { ... }
func TestCreateUser_InvalidInput(t *testing.T) { ... }

func TestGetUserByID_Found(t *testing.T) { ... }
func TestGetUserByID_NotFound(t *testing.T) { ... }

func TestValidateEmail_ValidFormat(t *testing.T) { ... }
func TestValidateEmail_InvalidFormat(t *testing.T) { ... }
func TestValidateEmail_EmptyString(t *testing.T) { ... }
```

### Method Tests

```go
// Pattern: Test{Type}_{Method}_{Scenario}

func TestUserService_Create_Success(t *testing.T) { ... }
func TestUserService_Create_ValidationError(t *testing.T) { ... }

func TestUserRepository_GetByID_Found(t *testing.T) { ... }
func TestUserRepository_GetByID_NotFound(t *testing.T) { ... }
```

---

## 3. AAA Pattern

### Arrange-Act-Assert

```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    repo := NewMockUserRepository()
    repo.On("Create", mock.Anything, mock.AnythingOfType("*User")).Return(nil)

    service := NewUserService(repo)

    input := CreateUserRequest{
        Name:  "John Doe",
        Email: "john@example.com",
    }

    // Act
    user, err := service.CreateUser(context.Background(), input)

    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, user.ID)
    assert.Equal(t, "John Doe", user.Name)
    assert.Equal(t, "john@example.com", user.Email)

    repo.AssertExpectations(t)
}
```

---

## 4. Using testify

### Assertions

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // assert - continues on failure
    assert.Equal(t, expected, actual)
    assert.NotNil(t, obj)
    assert.True(t, condition)
    assert.Contains(t, slice, element)
    assert.Len(t, slice, 3)

    // require - stops on failure
    require.NoError(t, err)      // must pass to continue
    require.NotNil(t, result)    // must pass to continue
}

// Comparison examples
assert.Equal(t, 123, num)                    // equality
assert.NotEqual(t, 0, num)                   // inequality
assert.Greater(t, 10, num)                   // num > 10
assert.GreaterOrEqual(t, 10, num)            // num >= 10
assert.Less(t, 5, num)                       // num < 5
assert.Empty(t, slice)                       // len == 0
assert.NotEmpty(t, slice)                    // len > 0
assert.Nil(t, ptr)                           // ptr == nil
assert.NotNil(t, ptr)                        // ptr != nil
assert.Error(t, err)                         // err != nil
assert.NoError(t, err)                       // err == nil
assert.ErrorIs(t, err, ErrNotFound)          // errors.Is
assert.ErrorContains(t, err, "not found")    // error message
```

### Mocking

```go
import "github.com/stretchr/testify/mock"

// Define mock
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

// Use in tests
func TestUserService_GetUser(t *testing.T) {
    // Setup mock
    repo := new(MockUserRepository)
    repo.On("GetByID", mock.Anything, "user-123").Return(&User{
        ID:    "user-123",
        Name:  "John",
        Email: "john@example.com",
    }, nil)

    service := NewUserService(repo)

    // Execute
    user, err := service.GetUser(context.Background(), "user-123")

    // Verify
    require.NoError(t, err)
    assert.Equal(t, "John", user.Name)

    // Verify mock was called correctly
    repo.AssertExpectations(t)
    repo.AssertCalled(t, "GetByID", mock.Anything, "user-123")
}

// Mock error case
func TestUserService_GetUser_NotFound(t *testing.T) {
    repo := new(MockUserRepository)
    repo.On("GetByID", mock.Anything, "invalid-id").Return(nil, ErrNotFound)

    service := NewUserService(repo)

    user, err := service.GetUser(context.Background(), "invalid-id")

    assert.Nil(t, user)
    assert.ErrorIs(t, err, ErrNotFound)
}
```

---

## 5. Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input int
    }{
        {"case1", 1},
        {"case2", 2},
        {"case3", 3},
    }

    for _, tt := range tests {
        tt := tt  // capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // run in parallel

            result := Process(tt.input)
            assert.NotNil(t, result)
        })
    }
}

// Note: Don't use t.Parallel() when tests share state
```

---

## 6. Test Helpers

### Setup/Teardown

```go
func TestMain(m *testing.M) {
    // Setup before all tests
    setupTestDB()

    code := m.Run()

    // Teardown after all tests
    teardownTestDB()

    os.Exit(code)
}

// Per-test setup
func setupTest(t *testing.T) (*Service, func()) {
    t.Helper()

    db := setupTestDB()
    service := NewService(db)

    cleanup := func() {
        db.Close()
    }

    return service, cleanup
}

func TestService(t *testing.T) {
    service, cleanup := setupTest(t)
    defer cleanup()

    // test logic
}
```

### Test Fixtures

```go
// testdata/users/valid.json
// testdata/users/invalid.json

func loadFixture(t *testing.T, name string) []byte {
    t.Helper()

    data, err := os.ReadFile(filepath.Join("testdata", name))
    require.NoError(t, err)

    return data
}

func TestParseUser_FromFixture(t *testing.T) {
    data := loadFixture(t, "users/valid.json")

    user, err := ParseUser(data)

    require.NoError(t, err)
    assert.Equal(t, "John", user.Name)
}
```

---

## 7. HTTP Handler Tests

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetUserHandler(t *testing.T) {
    // Setup
    service := new(MockUserService)
    service.On("GetUser", mock.Anything, "user-123").Return(&User{
        ID:   "user-123",
        Name: "John",
    }, nil)

    handler := NewHandler(service)

    // Create request
    req := httptest.NewRequest(http.MethodGet, "/users/user-123", nil)
    rec := httptest.NewRecorder()

    // Execute
    handler.GetUser(rec, req)

    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)

    var response User
    err := json.NewDecoder(rec.Body).Decode(&response)
    require.NoError(t, err)
    assert.Equal(t, "John", response.Name)
}

func TestCreateUserHandler(t *testing.T) {
    service := new(MockUserService)
    service.On("CreateUser", mock.Anything, mock.AnythingOfType("*CreateUserRequest")).Return(&User{
        ID:    "new-id",
        Name:  "John",
        Email: "john@example.com",
    }, nil)

    handler := NewHandler(service)

    body := `{"name": "John", "email": "john@example.com"}`
    req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.CreateUser(rec, req)

    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

---

## 8. Context with Timeout

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := SlowOperation(ctx)

    require.NoError(t, err)
    assert.NotNil(t, result)
}

func TestTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    time.Sleep(10 * time.Millisecond)  // ensure timeout

    _, err := SlowOperation(ctx)

    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
```

---

## 9. Benchmarks

```go
func BenchmarkProcess(b *testing.B) {
    data := generateTestData(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}

func BenchmarkProcess_Parallel(b *testing.B) {
    data := generateTestData(1000)

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            Process(data)
        }
    })
}

// Run: go test -bench=. -benchmem
```

---

## Quick Reference

| Task | Code |
|------|------|
| Run tests | `go test ./...` |
| Run with verbose | `go test -v ./...` |
| Run specific test | `go test -run TestName ./...` |
| Run with coverage | `go test -cover ./...` |
| Generate coverage report | `go test -coverprofile=coverage.out ./...` |
| View coverage | `go tool cover -html=coverage.out` |
| Run benchmarks | `go test -bench=. -benchmem ./...` |
| Race detection | `go test -race ./...` |

---

## Related Knowledge

- [01-go-fundamentals.md](./01-go-fundamentals.md) - Basic patterns
- [02-error-patterns.md](./02-error-patterns.md) - Testing error cases
