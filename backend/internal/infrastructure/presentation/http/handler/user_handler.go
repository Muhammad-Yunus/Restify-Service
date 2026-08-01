package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// UserHandler handles user HTTP requests.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// @Summary		Get user by ID
// @Description	Get a user by their ID
// @Tags			users
// @Produce		json
// @Param			id	path		string	true	"User ID"	format(uuid)
// @Success		200	{object}	map[string]interface{}
// @Failure		400	{object}	map[string]interface{}
// @Failure		404	{object}	map[string]interface{}
// @Router			/api/v1/users/:id [get]
// @Security		BearerAuth
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

// @Summary		List users
// @Description	List all users with pagination
// @Tags			users
// @Produce		json
// @Param			page		query		int	false	"Page number"	default(1)
// @Param			page_size	query		int	false	"Page size"		default(20)
// @Success		200			{object}	map[string]interface{}
// @Failure		500			{object}	map[string]interface{}
// @Router			/api/v1/users [get]
// @Security		BearerAuth
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       toUserListDTO(users),
		"pagination": gin.H{"page": page, "page_size": pageSize, "total": total},
	})
}

// @Summary		Update user
// @Description	Update user information
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"User ID"	format(uuid)
// @Param			body	body		object	true	"User updates"
// @Success		200		{object}	map[string]interface{}
// @Failure		400		{object}	map[string]interface{}
// @Failure		404		{object}	map[string]interface{}
// @Router			/api/v1/users/:id [patch]
// @Security		BearerAuth
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

// @Summary		Delete user
// @Description	Delete a user by ID (admin only)
// @Tags			users
// @Produce		json
// @Param			id	path		string	true	"User ID"	format(uuid)
// @Success		200	{object}	map[string]string
// @Failure		400	{object}	map[string]interface{}
// @Failure		404	{object}	map[string]interface{}
// @Router			/api/v1/users/:id [delete]
// @Security		BearerAuth
// @Security		AdminRole
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
