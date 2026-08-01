# Epic 08 — User Service & Repository

**Goal:** Implement user repository (GORM-based) and user application service with full CRUD operations.
**Dependencies:** Epic 04 (User entity), Epic 05 (Repository interfaces), Epic 06 (DB adapter)
**Commit:** `feat: implement user repository and service with full CRUD`

---

## Step 08.01 — User Repository Implementation

**Build:** Create `backend/internal/application/repository/user_repository.go`:

```go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
    "gorm.io/gorm"
)

// UserRepositoryImpl implements the repository.UserRepository interface.
type UserRepositoryImpl struct {
    db *gorm.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db repository.DB) repository.UserRepository {
    return &UserRepositoryImpl{db: gormDB(db)}
}

func gormDB(db repository.DB) *gorm.DB {
    // Access the GORM instance from the database infrastructure layer
    return db.DB() // will be adjusted when DB interface exposes GORM
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
    if user.ID == uuid.Nil {
        user.ID = uuid.New()
    }
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    return nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).
        Preload("Roles").
        First(&user, "id = ?", id).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find user by ID %s: %w", id, err)
    }
    return &user, nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    err := r.db.WithContext(ctx).
        Preload("Roles").
        First(&user, "email = ?", email).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("find user by email %s: %w", email, err)
    }
    return &user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
    if err := r.db.WithContext(ctx).Model(user).Updates(map[string]any{
        "full_name":    user.FullName,
        "is_active":    user.IsActive,
    }).Error; err != nil {
        return fmt.Errorf("update user %s: %w", user.ID, err)
    }
    return nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
    result := r.db.WithContext(ctx).Model(&entity.User{}).
        Where("id = ?", id).
        Update("is_active", false)

    if result.Error != nil {
        return fmt.Errorf("delete user %s: %w", id, result.Error)
    }
    if result.RowsAffected == 0 {
        return entity.ErrNotFound
    }
    return nil
}

func (r *UserRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error) {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }

    offset := (page - 1) * pageSize

    var users []*entity.User
    var total int64

    if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count users: %w", err)
    }

    if err := r.db.WithContext(ctx).
        Preload("Roles").
        Offset(offset).Limit(pageSize).
        Find(&users).Error; err != nil {
        return nil, 0, fmt.Errorf("list users: %w", err)
    }

    return users, int(total), nil
}

func (r *UserRepositoryImpl) CountActive(ctx context.Context) (int, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&entity.User{}).
        Where("is_active = ?", true).
        Count(&count).Error; err != nil {
        return 0, fmt.Errorf("count active users: %w", err)
    }
    return int(count), nil
}
```

**Test cases:**
- [ ] Unit: `Create()` inserts user with generated UUID
- [ ] Unit: `FindByID()` returns user when exists
- [ ] Unit: `FindByID()` returns ErrNotFound when missing
- [ ] Unit: `FindByEmail()` returns user when exists
- [ ] Unit: `Update()` updates only specified fields
- [ ] Unit: `Delete()` soft-deletes user
- [ ] Unit: `List()` returns paginated results
- [ ] Unit: `CountActive()` counts only active users
- [ ] Integration: Full CRUD cycle with test database

---

## Step 08.02 — User Application Service

**Build:** Create `backend/internal/application/service/user_service.go`:

```go
package service

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// UserService implements the service.UserService interface.
type UserService struct {
    repo      repository.UserRepository
    authRepo  repository.RoleRepository
    logger    repository.Logger
}

// NewUserService creates a new user service.
func NewUserService(repo repository.UserRepository, authRepo repository.RoleRepository, logger repository.Logger) repository.UserService {
    return &UserService{
        repo:     repo,
        authRepo: authRepo,
        logger:   logger,
    }
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("get user by email %s: %w", email, err)
    }
    return user, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, updates map[string]any) (*entity.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if name, ok := updates["full_name"].(string); ok {
        user.FullName = &name
    }
    if active, ok := updates["is_active"].(bool); ok {
        user.IsActive = active
    }

    if err := user.Validate(); err != nil {
        return nil, fmt.Errorf("validate user: %w", err)
    }

    if err := s.repo.Update(ctx, user); err != nil {
        return nil, fmt.Errorf("update user %s: %w", id, err)
    }

    s.logger.Info(ctx, "user updated", "user_id", id)
    return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("delete user %s: %w", id, err)
    }
    s.logger.Info(ctx, "user deleted", "user_id", id)
    return nil
}

func (s *UserService) List(ctx context.Context, page, pageSize int) ([]*entity.User, int, error) {
    users, total, err := s.repo.List(ctx, page, pageSize)
    if err != nil {
        return nil, 0, fmt.Errorf("list users: %w", err)
    }
    return users, total, nil
}

// AssignRole assigns a role to a user.
func (s *UserService) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
    return s.authRepo.AssignUser(ctx, userID, roleID)
}

// RemoveRole removes a role from a user.
func (s *UserService) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
    return s.authRepo.RemoveUser(ctx, userID, roleID)
}
```

**Test cases:**
- [ ] Unit: `GetByID()` returns user
- [ ] Unit: `GetByID()` returns error when user not found
- [ ] Unit: `Update()` updates full_name
- [ ] Unit: `Update()` returns validation error for invalid data
- [ ] Unit: `Delete()` soft-deletes user
- [ ] Unit: `List()` returns paginated users
- [ ] Unit: `AssignRole()` assigns role to user
- [ ] Unit: `RemoveRole()` removes role from user

---

## Step 08.03 — User HTTP Handler

**Build:** Create `backend/internal/presentation/http/handler/user_handler.go`:

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// UserHandler handles user HTTP requests.
type UserHandler struct {
    userService service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

// GetByID handles GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "user ID must be a UUID"))
        return
    }

    user, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "User not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toUserDTO(user))
}

// List handles GET /api/v1/users
func (h *UserHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

    users, total, err := h.userService.List(c.Request.Context(), page, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":      toUserListDTO(users),
        "pagination": gin.H{"page": page, "page_size": pageSize, "total": total},
    })
}

// Update handles PATCH /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "user ID must be a UUID"))
        return
    }

    var updates map[string]any
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid request body", err.Error()))
        return
    }

    user, err := h.userService.Update(c.Request.Context(), id, updates)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "User not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, toUserDTO(user))
}

// Delete handles DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Invalid ID", "user ID must be a UUID"))
        return
    }

    if err := h.userService.Delete(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusNotFound, dto.ProblemDetail(http.StatusNotFound, "User not found", err.Error()))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func toUserDTO(u *entity.User) gin.H {
    roles := make([]string, len(u.Roles))
    for i, r := range u.Roles {
        roles[i] = r.Name
    }
    return gin.H{
        "id":         u.ID.String(),
        "email":      u.Email,
        "full_name":  u.FullName,
        "is_active":  u.IsActive,
        "roles":      roles,
        "created_at": u.CreatedAt,
        "updated_at": u.UpdatedAt,
    }
}

func toUserListDTO(users []*entity.User) []gin.H {
    result := make([]gin.H, len(users))
    for i, u := range users {
        result[i] = toUserDTO(u)
    }
    return result
}
```

---

## Step 08.04 — Register User Routes

**Build:** Update router registration in `internal/presentation/http/router/routes.go`:

```go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/handler"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/middleware"
)

// RegisterUserRoutes registers user-related routes.
func RegisterUserRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware) {
    users := rg.Group("/users")
    {
        users.GET("", userHandler.List)
        users.GET("/:id", authMiddleware.RequireAuth(), userHandler.GetByID)
        users.PATCH("/:id", authMiddleware.RequireAuth(), userHandler.Update)
        users.DELETE("/:id", authMiddleware.RequireRole("administrator"), userHandler.Delete)
    }
}
```

**Test cases:**
- [ ] E2E: `GET /api/v1/users` returns paginated list
- [ ] E2E: `GET /api/v1/users/:id` returns user details
- [ ] E2E: `PATCH /api/v1/users/:id` updates user
- [ ] E2E: `DELETE /api/v1/users/:id` requires administrator role
- [ ] E2E: Unauthenticated requests to protected routes return 401

---

## Commit Instruction

```bash
git add .
git commit -m "feat: implement user repository, service, and HTTP handlers"
```
