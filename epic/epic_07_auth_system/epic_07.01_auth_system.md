# Epic 07 — Authentication System

**Goal:** Implement JWT authentication with HS256, bcrypt password hashing, token management, and auth middleware.
**Dependencies:** Epic 03 (Database), Epic 04 (User entity), Epic 06 (Logger, Cache adapters)
**Commit:** `feat: add JWT authentication system with bcrypt`

---

## Step 07.01 — JWT Token Service

**Build:** Create `backend/internal/infrastructure/auth/jwt.go`:

```go
package auth

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// JWTService handles token generation and validation.
type JWTService struct {
    secret     []byte
    expiration time.Duration
}

// NewJWTService creates a new JWT service.
func NewJWTService(secret string, expiration time.Duration) *JWTService {
    // Validate secret length
    if len(secret) < 32 {
        // Generate a random secret if too short
        b := make([]byte, 32)
        rand.Read(b)
        secret = base64.URLEncoding.EncodeToString(b)
    }
    return &JWTService{
        secret:     []byte(secret),
        expiration: expiration,
    }
}

// GenerateAccessToken creates a new access token for a user.
func (j *JWTService) GenerateAccessToken(userID string, email string, roles []string) (string, error) {
    claims := jwt.MapClaims{
        "sub":  userID,
        "email": email,
        "roles": roles,
        "type":  "access",
        "iat":   time.Now().Unix(),
        "exp":   time.Now().Add(j.expiration).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(j.secret)
    if err != nil {
        return "", fmt.Errorf("sign access token: %w", err)
    }
    return signed, nil
}

// GenerateRefreshToken creates a new refresh token.
func (j *JWTService) GenerateRefreshToken(userID string) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "type": "refresh",
        "iat": time.Now().Unix(),
        "exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(j.secret)
    if err != nil {
        return "", fmt.Errorf("sign refresh token: %w", err)
    }
    return signed, nil
}

// ParseToken validates and parses a JWT token.
func (j *JWTService) ParseToken(tokenString string) (jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return j.secret, nil
    })
    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    return claims, nil
}

// ExtractUserID extracts the user ID from a token.
func (j *JWTService) ExtractUserID(tokenString string) (string, error) {
    claims, err := j.ParseToken(tokenString)
    if err != nil {
        return "", err
    }
    sub, ok := claims["sub"].(string)
    if !ok {
        return "", fmt.Errorf("user ID not found in token")
    }
    return sub, nil
}

// ExtractRoles extracts roles from a token.
func (j *JWTService) ExtractRoles(tokenString string) ([]string, error) {
    claims, err := j.ParseToken(tokenString)
    if err != nil {
        return nil, err
    }
    rolesRaw, ok := claims["roles"]
    if !ok {
        return []string{}, nil
    }
    roles, ok := rolesRaw.([]any)
    if !ok {
        return nil, fmt.Errorf("roles has invalid type")
    }
    result := make([]string, len(roles))
    for i, r := range roles {
        result[i], ok = r.(string)
        if !ok {
            return nil, fmt.Errorf("role has invalid type")
        }
    }
    return result, nil
}
```

**Test cases:**
- [ ] Unit: `GenerateAccessToken()` produces valid JWT
- [ ] Unit: `ParseToken()` validates correct token
- [ ] Unit: `ParseToken()` rejects expired token
- [ ] Unit: `ParseToken()` rejects tampered token
- [ ] Unit: `ExtractUserID()` returns correct user ID
- [ ] Unit: `ExtractRoles()` returns correct roles
- [ ] Unit: Short secret auto-generates random 32-byte key
- [ ] Unit: Refresh token has 7-day expiration

---

## Step 07.02 — Password Hashing Service

**Build:** Create `backend/internal/infrastructure/auth/password.go`:

```go
package auth

import "golang.org/x/crypto/bcrypt"

// PasswordService handles password hashing and verification.
type PasswordService struct {
    cost int
}

// NewPasswordService creates a new password service.
func NewPasswordService(cost int) *PasswordService {
    if cost < 10 {
        cost = 12 // default minimum
    }
    return &PasswordService{cost: cost}
}

// Hash generates a bcrypt hash from plaintext password.
func (p *PasswordService) Hash(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(bytes), nil
}

// Verify checks if plaintext password matches hash.
func (p *PasswordService) Verify(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// NeedsRehash checks if the hash cost needs to be upgraded.
func (p *PasswordService) NeedsRehash(hash string) bool {
    cost, err := bcrypt.Cost([]byte(hash))
    if err != nil {
        return false
    }
    return cost < p.cost
}
```

**Test cases:**
- [ ] Unit: `Hash()` produces non-empty hash
- [ ] Unit: `Verify()` returns true for matching password
- [ ] Unit: `Verify()` returns false for wrong password
- [ ] Unit: `NeedsRehash()` returns true when cost is too low
- [ ] Unit: Default cost is 12

---

## Step 07.03 — Token Blacklist Service (Redis-backed)

**Build:** Create `backend/internal/infrastructure/auth/blacklist.go`:

```go
package auth

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// TokenBlacklist implements the repository.TokenBlacklistRepository interface.
type TokenBlacklist struct {
    cache repository.Cache
    prefix string
}

// NewTokenBlacklist creates a new token blacklist service.
func NewTokenBlacklist(cache repository.Cache) *TokenBlacklist {
    return &TokenBlacklist{
        cache:  cache,
        prefix: "jwt:blacklist:",
    }
}

func (tb *TokenBlacklist) hashToken(token string) string {
    h := sha256.Sum256([]byte(token))
    return hex.EncodeToString(h[:])
}

func (tb *TokenBlacklist) key(tokenHash string) string {
    return tb.prefix + tokenHash
}

func (tb *TokenBlacklist) Add(ctx context.Context, token string, expiresAt time.Time) error {
    tokenHash := tb.hashToken(token)
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        return nil // token already expired
    }
    return tb.cache.Set(ctx, tb.key(tokenHash), "1", ttl)
}

func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
    exists, err := tb.cache.Exists(ctx, tb.key(tb.hashToken(token)))
    if err != nil {
        return false, fmt.Errorf("check blacklist: %w", err)
    }
    return exists, nil
}

func (tb *TokenBlacklist) Cleanup(ctx context.Context) (int64, error) {
    // Redis automatically expires keys; this is a no-op in practice.
    // Kept for interface compliance and future manual cleanup if needed.
    return 0, nil
}

// Compile-time check.
var _ repository.TokenBlacklistRepository = (*TokenBlacklist)(nil)
```

**Test cases:**
- [ ] Unit: `Add()` stores hashed token in cache
- [ ] Unit: `IsBlacklisted()` returns true for blacklisted token
- [ ] Unit: `IsBlacklisted()` returns false for valid token
- [ ] Unit: `hashToken()` produces consistent SHA256 hash
- [ ] Unit: Expired tokens are not added

---

## Step 07.04 — Auth Middleware

**Build:** Create `backend/internal/infrastructure/presentation/http/middleware/auth.go`:

```go
package middleware

import (
    "net/http"
    "strings"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/infrastructure/auth"
)

// AuthMiddleware provides JWT authentication.
type AuthMiddleware struct {
    jwtService     *auth.JWTService
    blacklist      *auth.TokenBlacklist
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(jwtService *auth.JWTService, blacklist *auth.TokenBlacklist) *AuthMiddleware {
    return &AuthMiddleware{
        jwtService:  jwtService,
        blacklist:   blacklist,
    }
}

// RequireAuth middleware that validates JWT and sets user context.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            c.Abort()
            return
        }

        // Check blacklist
        if blacklisted, err := m.blacklist.IsBlacklisted(c.Request.Context(), tokenString); err != nil || blacklisted {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "token is revoked"})
            c.Abort()
            return
        }

        claims, err := m.jwtService.ParseToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("user_id", claims["sub"])
        c.Set("email", claims["email"])
        c.Set("token", tokenString)

        // Extract and set roles
        if rolesRaw, ok := claims["roles"]; ok && rolesRaw != nil {
            if roles, ok := rolesRaw.([]any); ok {
                roleList := make([]string, len(roles))
                for i, r := range roles {
                    if s, ok := r.(string); ok {
                        roleList[i] = s
                    }
                }
                c.Set("roles", roleList)
            }
        }

        c.Next()
    }
}

// RequireRole middleware that checks user has at least one of the required roles.
func (m *AuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        rolesRaw, exists := c.Get("roles")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }

        userRoles, ok := rolesRaw.([]string)
        if !ok {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }

        roleSet := make(map[string]bool, len(userRoles))
        for _, r := range userRoles {
            roleSet[r] = true
        }

        for _, required := range allowedRoles {
            if roleSet[required] {
                c.Next()
                return
            }
        }

        c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
        c.Abort()
    }
}

// OptionalAuth middleware that sets user context if token is present but doesn't require it.
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Next()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.Next()
            return
        }

        claims, err := m.jwtService.ParseToken(tokenString)
        if err != nil {
            c.Next()
            return
        }

        c.Set("user_id", claims["sub"])
        c.Set("email", claims["email"])
        c.Set("token", tokenString)
        c.Next()
    }
}
```

**Test cases:**
- [ ] Unit: `RequireAuth()` rejects request without Authorization header
- [ ] Unit: `RequireAuth()` rejects invalid token format
- [ ] Unit: `RequireAuth()` accepts valid token and sets user context
- [ ] Unit: `RequireAuth()` rejects blacklisted token
- [ ] Unit: `RequireRole()` allows request with matching role
- [ ] Unit: `RequireRole()` rejects request without matching role
- [ ] Unit: `OptionalAuth()` allows request without token
- [ ] Unit: `OptionalAuth()` sets context when valid token present

---

## Step 07.05 — Auth DTOs and Handler

**Build:** Create `backend/internal/presentation/http/dto/auth.go`:

```go
package dto

import "github.com/go-playground/validator/v10"

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    FullName string `json:"full_name,omitempty" validate:"max=255"`
}

func (r *RegisterRequest) Validate() error {
    return validator.New().Struct(r)
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

func (r *LoginRequest) Validate() error {
    return validator.New().Struct(r)
}

// RefreshRequest is the request body for token refresh.
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r *RefreshRequest) Validate() error {
    return validator.New().Struct(r)
}

// AuthResponse is the response body for auth operations.
type AuthResponse struct {
    User         *UserResponse `json:"user"`
    AccessToken  string        `json:"access_token"`
    RefreshToken string        `json:"refresh_token"`
    ExpiresIn    int           `json:"expires_in"`
}

// UserResponse is a sanitized user representation.
type UserResponse struct {
    ID       string   `json:"id"`
    Email    string   `json:"email"`
    FullName *string  `json:"full_name,omitempty"`
    Roles    []string `json:"roles"`
}
```

**Build:** Create `backend/internal/presentation/http/handler/auth_handler.go`:

```go
package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/muhammadyunus/ForgeBase/internal/application/service"
    "github.com/muhammadyunus/ForgeBase/internal/domain/entity"
    "github.com/muhammadyunus/ForgeBase/internal/infrastructure/auth"
    "github.com/muhammadyunus/ForgeBase/internal/presentation/http/dto"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
    authService service.AuthService
    jwtService  *auth.JWTService
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService service.AuthService, jwtService *auth.JWTService) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        jwtService:  jwtService,
    }
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
    var req dto.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    if err := req.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
    if err != nil {
        c.JSON(http.StatusConflict, ProblemDetail(http.StatusConflict, "Email already registered", err.Error()))
        return
    }

    accessToken, err := h.jwtService.GenerateAccessToken(user.ID.String(), user.Email, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ProblemDetail(http.StatusInternalServerError, "Token generation failed", err.Error()))
        return
    }
    refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.String())
    if err != nil {
        c.JSON(http.StatusInternalServerError, ProblemDetail(http.StatusInternalServerError, "Token generation failed", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, dto.AuthResponse{
        User:         toUserResponse(user),
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int(24 * time.Hour.Seconds()),
    })
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    result, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, ProblemDetail(http.StatusUnauthorized, "Invalid credentials", err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.AuthResponse{
        User:         toUserResponse(result.User),
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
        ExpiresIn:    result.ExpiresIn,
    })
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
    var req dto.RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))
        return
    }

    result, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, ProblemDetail(http.StatusUnauthorized, "Invalid refresh token", err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.AuthResponse{
        User:         toUserResponse(result.User),
        AccessToken:  result.AccessToken,
        RefreshToken: result.RefreshToken,
        ExpiresIn:    result.ExpiresIn,
    })
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
    token := c.GetString("token")
    if err := h.authService.Logout(c.Request.Context(), token); err != nil {
        c.JSON(http.StatusInternalServerError, ProblemDetail(http.StatusInternalServerError, "Logout failed", err.Error()))
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func toUserResponse(u *entity.User) *dto.UserResponse {
    roles := make([]string, len(u.Roles))
    for i, r := range u.Roles {
        roles[i] = r.Name
    }
    return &dto.UserResponse{
        ID:       u.ID.String(),
        Email:    u.Email,
        FullName: u.FullName,
        Roles:    roles,
    }
}
```

**Test cases:**
- [ ] Unit: `Register()` creates user and returns token
- [ ] Unit: `Register()` returns 409 when email exists
- [ ] Unit: `Register()` validates input fields
- [ ] Unit: `Login()` returns tokens for valid credentials
- [ ] Unit: `Login()` returns 401 for invalid credentials
- [ ] Unit: `Refresh()` returns new tokens for valid refresh token
- [ ] Unit: `Logout()` blacklists the token

---

## Step 07.06 — RFC 7807 Problem Detail Helper

**Build:** Create `backend/internal/presentation/http/dto/problem_detail.go`:

```go
package dto

import (
    "fmt"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

// ProblemDetail creates an RFC 7807 Problem Details response.
func ProblemDetail(status int, title string, detail string) gin.H {
    return gin.H{
        "type":    fmt.Sprintf("https://ForgeBase.api/errors/%s", strings.ToLower(strings.ReplaceAll(title, " ", "-"))),
        "title":   title,
        "status":  status,
        "detail":  detail,
        "instance": http.StatusText(status),
    }
}

// ProblemDetailWithPath creates an RFC 7807 Problem Details response with a specific path.
func ProblemDetailWithPath(status int, title string, detail string, path string) gin.H {
    resp := ProblemDetail(status, title, detail)
    resp["instance"] = path
    return resp
}

// ProblemDetailWithErrors creates a Problem Details response with validation errors.
func ProblemDetailWithErrors(status int, title string, detail string, errors []FieldError) gin.H {
    resp := ProblemDetail(status, title, detail)
    resp["errors"] = errors
    return resp
}

// FieldError represents a single validation error.
type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

**Test cases:**
- [ ] Unit: `ProblemDetail()` returns correct JSON structure
- [ ] Unit: `ProblemDetailWithErrors()` includes errors array
- [ ] Unit: Type URL follows RFC 7807 format

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add JWT authentication system with bcrypt, token blacklist, and auth middleware"
```
