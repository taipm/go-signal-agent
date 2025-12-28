---
stepNumber: 5
nextStep: './step-06-review-loop.md'
agent: test-agent
hasBreakpoint: false
---

# Step 05: Test Creation

## STEP GOAL

Test Agent creates comprehensive tests for the implemented code, following table-driven patterns and achieving good coverage.

## AGENT ACTIVATION

Load persona từ `../agents/test-agent.md`

Input context:
- Code files từ step 04
- Spec từ step 02 (acceptance criteria)
- Architecture từ step 03 (interfaces)

## EXECUTION SEQUENCE

### 1. Test Agent Introduction

```
[Test Agent]

Analyzing code để tạo test suite...

Test strategy:
1. Unit tests cho Service layer
2. Unit tests cho Handler layer
3. Integration tests (optional)

Target: 80%+ coverage
```

### 2. Create Mock Interfaces

```go
// internal/repo/mock_repo.go
package repo

import (
    "context"
    "github.com/stretchr/testify/mock"
    "{module}/internal/model"
)

type Mock{Entity}Repo struct {
    mock.Mock
}

func (m *Mock{Entity}Repo) GetByID(ctx context.Context, id string) (*model.{Entity}, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.{Entity}), args.Error(1)
}

func (m *Mock{Entity}Repo) Create(ctx context.Context, entity *model.{Entity}) error {
    args := m.Called(ctx, entity)
    return args.Error(0)
}
```

### 3. Create Service Tests

```go
// internal/service/{entity}_service_test.go
package service_test

import (
    "context"
    "log/slog"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "{module}/internal/model"
    "{module}/internal/repo"
    "{module}/internal/service"
)

func TestService_GetByID(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name      string
        id        string
        mockSetup func(*repo.Mock{Entity}Repo)
        want      *model.{Entity}
        wantErr   error
    }{
        {
            name: "success",
            id:   "123",
            mockSetup: func(m *repo.Mock{Entity}Repo) {
                m.On("GetByID", mock.Anything, "123").
                    Return(&model.{Entity}{ID: "123"}, nil)
            },
            want: &model.{Entity}{ID: "123"},
        },
        {
            name: "not found",
            id:   "invalid",
            mockSetup: func(m *repo.Mock{Entity}Repo) {
                m.On("GetByID", mock.Anything, "invalid").
                    Return(nil, model.ErrNotFound)
            },
            wantErr: model.ErrNotFound,
        },
        {
            name:      "empty id",
            id:        "",
            mockSetup: func(m *repo.Mock{Entity}Repo) {},
            wantErr:   model.ErrInvalidInput,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            // Arrange
            mockRepo := new(repo.Mock{Entity}Repo)
            tt.mockSetup(mockRepo)
            svc := service.New{Entity}Service(mockRepo, slog.Default())

            // Act
            got, err := svc.GetByID(context.Background(), tt.id)

            // Assert
            if tt.wantErr != nil {
                require.Error(t, err)
                assert.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want.ID, got.ID)
        })
    }
}
```

### 4. Create Handler Tests

```go
// internal/handler/{entity}_handler_test.go
package handler_test

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "{module}/internal/handler"
    "{module}/internal/model"
)

func TestHandler_Get(t *testing.T) {
    tests := []struct {
        name       string
        id         string
        mockSetup  func(*MockService)
        wantStatus int
        wantBody   *model.{Entity}
    }{
        {
            name: "success",
            id:   "123",
            mockSetup: func(m *MockService) {
                m.On("GetByID", mock.Anything, "123").
                    Return(&model.{Entity}{ID: "123"}, nil)
            },
            wantStatus: http.StatusOK,
            wantBody:   &model.{Entity}{ID: "123"},
        },
        {
            name:       "missing id",
            id:         "",
            mockSetup:  func(m *MockService) {},
            wantStatus: http.StatusBadRequest,
        },
        {
            name: "not found",
            id:   "invalid",
            mockSetup: func(m *MockService) {
                m.On("GetByID", mock.Anything, "invalid").
                    Return(nil, model.ErrNotFound)
            },
            wantStatus: http.StatusNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockSvc := new(MockService)
            tt.mockSetup(mockSvc)
            h := handler.New{Entity}Handler(mockSvc, slog.Default())

            req := httptest.NewRequest("GET", "/{entities}/"+tt.id, nil)
            req.SetPathValue("id", tt.id)
            rec := httptest.NewRecorder()

            // Act
            h.Get(rec, req)

            // Assert
            assert.Equal(t, tt.wantStatus, rec.Code)

            if tt.wantBody != nil {
                var got model.{Entity}
                json.NewDecoder(rec.Body).Decode(&got)
                assert.Equal(t, tt.wantBody.ID, got.ID)
            }
        })
    }
}
```

### 5. Run Tests

```bash
# Run all tests
go test ./... -v -race

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Coverage report
go tool cover -html=coverage.out -o coverage.html
```

### 6. Report Results

```
[Test Agent]

Test suite created:

Files:
- internal/repo/mock_repo.go
- internal/service/{entity}_service_test.go
- internal/handler/{entity}_handler_test.go

Test Results:
- Total tests: X
- Passed: X
- Failed: 0

Coverage: XX%

Ready for Review phase.
```

## OUTPUT

```yaml
outputs:
  test_files:
    - path: "internal/repo/mock_repo.go"
    - path: "internal/service/{entity}_service_test.go"
    - path: "internal/handler/{entity}_handler_test.go"
  test_results:
    total: 10
    passed: 10
    failed: 0
  coverage: 85
```

## SUCCESS CRITERIA

- [ ] Mock interfaces created
- [ ] Service tests with table-driven pattern
- [ ] Handler tests
- [ ] All tests pass
- [ ] Coverage >= 80%
- [ ] Ready for Review phase

## NEXT STEP

Load `./step-06-review-loop.md`
