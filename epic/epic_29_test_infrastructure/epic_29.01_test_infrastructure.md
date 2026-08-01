# Epic 29 — Test Infrastructure

**Goal:** Set up test infrastructure with testcontainers, mocks, and integration test suite.
**Dependencies:** All previous epics
**Commit:** `feat: add test infrastructure with testcontainers and mocks`

---

## Step 29.01 — Test Helpers

**Build:** Create `backend/test/helpers.go`:

```go
package test

import (
    "context"
    "fmt"
    "os"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB creates a test PostgreSQL container.
func SetupTestDB(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, string, func()) {
    t.Helper()

    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(10*time.Second),
        ),
    )
    if err != nil {
        t.Fatalf("failed to create postgres container: %v", err)
    }

    cleanup := func() {
        container.Terminate(ctx)
    }

    // Get connection string
    connString, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        cleanup()
        t.Fatalf("failed to get connection string: %v", err)
    }

    t.Cleanup(cleanup)
    return container, connString, cleanup
}

// SetupTestPool creates a pgxpool from test container.
func SetupTestPool(t *testing.T, ctx context.Context, connString string) *pgxpool.Pool {
    t.Helper()

    pool, err := pgxpool.NewWithConfig(ctx, &pgxpool.Config{
        ConnString: connString,
    })
    if err != nil {
        t.Fatalf("failed to create pool: %v", err)
    }

    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        t.Fatalf("failed to ping test DB: %v", err)
    }

    return pool
}

// LoadEnv loads .env.test file for test configuration.
func LoadEnvTest(t *testing.T) {
    t.Helper()
    if err := godotenv.Read(".env.test"); err != nil {
        // .env.test not required; use defaults
    }
}

// GetTestEnv gets environment variable or falls back to test default.
func GetTestEnv(key string, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
```

---

## Step 29.02 — Mock Generators

**Build:** Create `backend/test/mocks/generate.go`:

```go
package mocks

//go:generate mockgen -destination=../test/mocks/user_repository.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository UserRepository
//go:generate mockgen -destination=../test/mocks/workspace_repository.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository WorkspaceRepository
//go:generate mockgen -destination=../test/mocks/endpoint_repository.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository EndpointRepository
//go:generate mockgen -destination=../test/mocks/auth_service.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/service AuthService
//go:generate mockgen -destination=../test/mocks/cache.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository Cache
//go:generate mockgen -destination=../test/mocks/logger.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository Logger
```

---

## Step 29.03 — Unit Test Examples

**Build:** Create `backend/internal/application/service/user_service_test.go`:

```go
package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockUserRepository implements repository.UserRepository for testing.
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error) {
    args := m.Called(ctx, page, pageSize)
    return args.Get(0).([]*entity.User), args.GetInt(1), args.Error(2)
}

func (m *MockUserRepository) CountActive(ctx context.Context) (int, error) {
    args := m.Called(ctx)
    return args.GetInt(0), args.Error(1)
}

func TestUserService_GetByID(t *testing.T) {
    mockRepo := new(MockUserRepository)
    logger := new(MockLogger)
    svc := service.NewUserService(mockRepo, nil, logger)

    expectedUser := &entity.User{
        ID:    uuid.New(),
        Email: "test@example.com",
    }

    mockRepo.On("FindByID", mock.Anything, expectedUser.ID).Return(expectedUser, nil).Once()

    user, err := svc.GetByID(context.Background(), expectedUser.ID)
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, expectedUser.Email, user.Email)

    mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
    mockRepo := new(MockUserRepository)
    logger := new(MockLogger)
    svc := service.NewUserService(mockRepo, nil, logger)

    mockRepo.On("FindByID", mock.Anything, mock.Anything).Return((*entity.User)(nil), entity.ErrNotFound).Once()

    _, err := svc.GetByID(context.Background(), uuid.New())
    assert.ErrorIs(t, err, entity.ErrNotFound)

    mockRepo.AssertExpectations(t)
}
```

---

## Step 29.04 — Integration Test Examples

**Build:** Create `backend/test/integration/user_test.go`:

```go
package integration

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserAPI_CreateAndGet(t *testing.T) {
    ctx := context.Background()
    pool := setupTestDB(t, ctx)

    // Initialize app with test DB
    app := setupTestApp(t, ctx, pool)

    // Register user
    w := httptest.NewRecorder()
    body := `{"email":"test@example.com","password":"password123","full_name":"Test User"}`
    req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    app.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    assert.Contains(t, w.Body.String(), "access_token")

    // Get user
    w = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/api/v1/users/me", nil)
    req.Header.Set("Authorization", "Bearer "+extractToken(w.Body.String()))
    app.Router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "test@example.com")
}
```

---

## Step 29.05 — Test Makefile Targets

**Build:** Add to `Makefile`:

```makefile
test:
	go test ./... -race -count=1

test-unit:
	go test ./test/unit/... -coverprofile=coverage.out

test-integration:
	docker-compose -f test/docker-compose.test.yml up -d
	go test ./test/integration/... -race -count=1
	docker-compose -f test/docker-compose.test.yml down

test-coverage:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

generate-mocks:
	mockgen -destination=test/mocks/user_repository.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository UserRepository
	mockgen -destination=test/mocks/workspace_repository.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository WorkspaceRepository
	mockgen -destination=test/mocks/endpoint_repository.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/repository EndpointRepository
	mockgen -destination=test/mocks/auth_service.go -package=mocks github.com/muhammadyunus/ForgeBase/internal/domain/service AuthService
```

**Test cases:**
- [ ] Unit: All unit tests pass with mocks
- [ ] Integration: Full API flow with testcontainers
- [ ] Coverage: ≥ 80% code coverage
- [ ] Race: No race conditions detected

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add test infrastructure with testcontainers, mocks, and integration tests"
```
