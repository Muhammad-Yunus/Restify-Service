package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/auth"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
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

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))

		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))

		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		c.JSON(http.StatusConflict, dto.ProblemDetail(http.StatusConflict, "Email already registered", err.Error()))

		return
	}

	accessToken, err := h.jwtService.GenerateAccessToken(user.ID.String(), user.Email, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Token generation failed", err.Error()))

		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Token generation failed", err.Error()))

		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(24 * time.Hour.Seconds()),
	})
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))

		return
	}

	result, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ProblemDetail(http.StatusUnauthorized, "Invalid credentials", err.Error()))

		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		User:         toUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Validation Error", err.Error()))

		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ProblemDetail(http.StatusUnauthorized, "Invalid refresh token", err.Error()))

		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		User:         toUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString("token")
	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Logout failed", err.Error()))

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
