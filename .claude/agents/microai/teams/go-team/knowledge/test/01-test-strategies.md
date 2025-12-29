# Test Strategies - Test Agent Knowledge

**Version:** 1.0.0
**Agent:** Test Agent

---

## TL;DR

- Unit test mỗi layer độc lập
- Mock interfaces cho dependencies
- Table-driven tests cho multiple scenarios
- Integration tests cho critical paths
- Target 80%+ coverage

---

## 1. Test Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E╲         ← Few, slow, expensive
                 ╱──────╲
                ╱        ╲
               ╱Integration╲     ← Some, medium
              ╱────────────╲
             ╱              ╲
            ╱   Unit Tests   ╲   ← Many, fast, cheap
           ╱──────────────────╲
```

### Distribution

| Type | Coverage | Speed | Count |
|------|----------|-------|-------|
| Unit | 70% | Fast | Many |
| Integration | 20% | Medium | Some |
| E2E | 10% | Slow | Few |

---

## 2. Unit Tests by Layer

### Repository Tests

```go
// internal/repository/user_test.go

func TestUserRepository_Create(t *testing.T) {
    // Use test database or sqlmock
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := NewUserRepository(db)

    tests := []struct {
        name    string
        user    *model.User
        mockFn  func()
        wantErr bool
    }{
        {
            name: "success",
            user: &model.User{
                ID:    "user-123",
                Email: "test@example.com",
                Name:  "Test User",
            },
            mockFn: func() {
                mock.ExpectExec("INSERT INTO users").
                    WithArgs("user-123", "test@example.com", "Test User", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
                    WillReturnResult(sqlmock.NewResult(1, 1))
            },
            wantErr: false,
        },
        {
            name: "duplicate email",
            user: &model.User{
                ID:    "user-456",
                Email: "existing@example.com",
            },
            mockFn: func() {
                mock.ExpectExec("INSERT INTO users").
                    WillReturnError(&pq.Error{Code: "23505"})
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.mockFn()

            err := repo.Create(context.Background(), tt.user)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}
```

### Service Tests

```go
// internal/service/user_test.go

type mockUserRepository struct {
    mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     model.CreateUserRequest
        mockFn  func(*mockUserRepository)
        want    *model.User
        wantErr error
    }{
        {
            name: "success",
            req: model.CreateUserRequest{
                Email:    "new@example.com",
                Name:     "New User",
                Password: "password123",
            },
            mockFn: func(m *mockUserRepository) {
                m.On("GetByEmail", mock.Anything, "new@example.com").
                    Return(nil, model.ErrUserNotFound)
                m.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).
                    Return(nil)
            },
            want: &model.User{
                Email: "new@example.com",
                Name:  "New User",
            },
        },
        {
            name: "email already exists",
            req: model.CreateUserRequest{
                Email:    "existing@example.com",
                Name:     "Existing",
                Password: "password123",
            },
            mockFn: func(m *mockUserRepository) {
                m.On("GetByEmail", mock.Anything, "existing@example.com").
                    Return(&model.User{Email: "existing@example.com"}, nil)
            },
            wantErr: model.ErrEmailAlreadyUsed,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(mockUserRepository)
            tt.mockFn(mockRepo)

            logger := slog.New(slog.NewTextHandler(io.Discard, nil))
            service := NewUserService(mockRepo, nil, logger)

            user, err := service.CreateUser(context.Background(), tt.req)

            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want.Email, user.Email)
            assert.Equal(t, tt.want.Name, user.Name)
            assert.NotEmpty(t, user.ID)

            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Handler Tests

```go
// internal/handler/user_test.go

type mockUserService struct {
    mock.Mock
}

func (m *mockUserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func TestUserHandler_Create(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        mockFn     func(*mockUserService)
        wantStatus int
        wantBody   string
    }{
        {
            name: "success",
            body: `{"email":"test@example.com","name":"Test","password":"password123"}`,
            mockFn: func(m *mockUserService) {
                m.On("CreateUser", mock.Anything, mock.AnythingOfType("model.CreateUserRequest")).
                    Return(&model.User{
                        ID:    "user-123",
                        Email: "test@example.com",
                        Name:  "Test",
                    }, nil)
            },
            wantStatus: http.StatusCreated,
            wantBody:   `"id":"user-123"`,
        },
        {
            name:       "invalid json",
            body:       `{invalid}`,
            mockFn:     func(m *mockUserService) {},
            wantStatus: http.StatusBadRequest,
        },
        {
            name: "email exists",
            body: `{"email":"existing@example.com","name":"Test","password":"password123"}`,
            mockFn: func(m *mockUserService) {
                m.On("CreateUser", mock.Anything, mock.AnythingOfType("model.CreateUserRequest")).
                    Return(nil, model.ErrEmailAlreadyUsed)
            },
            wantStatus: http.StatusConflict,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := new(mockUserService)
            tt.mockFn(mockService)

            handler := NewUserHandler(mockService)

            req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            handler.Create(rec, req)

            assert.Equal(t, tt.wantStatus, rec.Code)
            if tt.wantBody != "" {
                assert.Contains(t, rec.Body.String(), tt.wantBody)
            }

            mockService.AssertExpectations(t)
        })
    }
}
```

---

## 3. Integration Tests

```go
// tests/integration/user_test.go

func TestUserFlow_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Create real dependencies
    repo := repository.NewUserRepository(db)
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    service := service.NewUserService(repo, nil, logger)
    handler := handler.NewUserHandler(service)

    // Setup router
    r := chi.NewRouter()
    r.Post("/users", handler.Create)
    r.Get("/users/{id}", handler.Get)

    server := httptest.NewServer(r)
    defer server.Close()

    // Test create user
    createBody := `{"email":"integration@test.com","name":"Integration Test","password":"password123"}`
    resp, err := http.Post(server.URL+"/users", "application/json", strings.NewReader(createBody))
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createResp model.CreateUserResponse
    json.NewDecoder(resp.Body).Decode(&createResp)
    resp.Body.Close()

    assert.NotEmpty(t, createResp.ID)
    assert.Equal(t, "integration@test.com", createResp.Email)

    // Test get user
    resp, err = http.Get(server.URL + "/users/" + createResp.ID)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    resp.Body.Close()

    // Test duplicate email
    resp, err = http.Post(server.URL+"/users", "application/json", strings.NewReader(createBody))
    require.NoError(t, err)
    assert.Equal(t, http.StatusConflict, resp.StatusCode)
    resp.Body.Close()
}

func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL not set")
    }

    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)

    // Run migrations
    runMigrations(db)

    // Cleanup after test
    t.Cleanup(func() {
        db.Exec("TRUNCATE users CASCADE")
    })

    return db
}
```

---

## 4. Test Coverage

### Coverage Commands

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View HTML report
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go tool cover -func=coverage.out | grep total

# Coverage for specific package
go test -coverprofile=coverage.out -covermode=atomic ./internal/service/...
```

### Coverage Targets

| Package | Min Coverage |
|---------|-------------|
| service | 85% |
| repository | 75% |
| handler | 80% |
| model | 60% |
| Overall | 80% |

---

## 5. Test Utilities

### Test Fixtures

```go
// tests/fixtures/users.go

package fixtures

var ValidUser = &model.User{
    ID:        "user-fixture-001",
    Email:     "fixture@test.com",
    Name:      "Fixture User",
    Status:    model.StatusActive,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

var ValidCreateRequest = model.CreateUserRequest{
    Email:    "new@test.com",
    Name:     "New User",
    Password: "SecurePass123!",
}

func NewUser(overrides ...func(*model.User)) *model.User {
    user := &model.User{
        ID:        uuid.New().String(),
        Email:     fmt.Sprintf("user-%d@test.com", time.Now().UnixNano()),
        Name:      "Test User",
        Status:    model.StatusActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    for _, override := range overrides {
        override(user)
    }
    return user
}

// Usage
user := fixtures.NewUser(func(u *model.User) {
    u.Status = model.StatusInactive
})
```

### Test Helpers

```go
// tests/helpers/helpers.go

package helpers

func RequireNoError(t *testing.T, err error, msg ...string) {
    t.Helper()
    if err != nil {
        if len(msg) > 0 {
            t.Fatalf("%s: %v", msg[0], err)
        }
        t.Fatalf("unexpected error: %v", err)
    }
}

func AssertJSONEqual(t *testing.T, expected, actual string) {
    t.Helper()

    var expectedJSON, actualJSON interface{}
    require.NoError(t, json.Unmarshal([]byte(expected), &expectedJSON))
    require.NoError(t, json.Unmarshal([]byte(actual), &actualJSON))

    assert.Equal(t, expectedJSON, actualJSON)
}

func ContextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
    t.Helper()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    t.Cleanup(cancel)
    return ctx
}
```

---

## 6. Test Organization

### File Structure

```
project/
├── internal/
│   ├── handler/
│   │   ├── user.go
│   │   └── user_test.go       # Unit tests
│   ├── service/
│   │   ├── user.go
│   │   └── user_test.go       # Unit tests
│   └── repository/
│       ├── user.go
│       └── user_test.go       # Unit tests with sqlmock
│
└── tests/
    ├── integration/
    │   └── user_test.go       # Integration tests
    ├── fixtures/
    │   └── users.go           # Test fixtures
    └── helpers/
        └── helpers.go         # Test utilities
```

### Test Tags

```go
// +build integration

package integration

// Run with: go test -tags=integration ./tests/integration/...
```

---

## Quick Reference

| Test Type | Mock? | DB? | Speed |
|-----------|-------|-----|-------|
| Handler | Yes | No | Fast |
| Service | Yes | No | Fast |
| Repository | sqlmock | No | Fast |
| Integration | No | Yes | Slow |

| Command | Purpose |
|---------|---------|
| `go test ./...` | Run all tests |
| `go test -v ./...` | Verbose output |
| `go test -run TestName` | Run specific test |
| `go test -cover` | Show coverage |
| `go test -race` | Race detection |
| `go test -short` | Skip slow tests |

---

## Related Knowledge

- [02-mocking-patterns.md](./02-mocking-patterns.md) - Mock strategies
- [../shared/04-testing-patterns.md](../shared/04-testing-patterns.md) - General patterns
