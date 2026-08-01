package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/auth"
)

// AuthMiddleware provides JWT authentication.
type AuthMiddleware struct {
	jwtService *auth.JWTService
	blacklist  *auth.TokenBlacklist
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(jwtService *auth.JWTService, blacklist *auth.TokenBlacklist) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		blacklist:  blacklist,
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
