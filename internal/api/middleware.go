package api

import (
	"net/http"
	"strings"

	"courier_managment_system_go/internal/domain"
	"courier_managment_system_go/internal/service"

	"github.com/gin-gonic/gin"
)

const userContextKey = "current_user"

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "forbidden"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "forbidden"})
			return
		}
		c.Set(userContextKey, service.UserFromClaims(claims))
		c.Next()
	}
}

func RequireRoles(roles ...domain.UserRole) gin.HandlerFunc {
	roleSet := make(map[domain.UserRole]struct{}, len(roles))
	for _, role := range roles {
		roleSet[role] = struct{}{}
	}
	return func(c *gin.Context) {
		user, ok := c.Get(userContextKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "forbidden"})
			return
		}
		authUser := user.(service.AuthUser)
		if len(roleSet) > 0 {
			if _, allowed := roleSet[authUser.Role]; !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "forbidden"})
				return
			}
		}
		c.Next()
	}
}

func MustUser(c *gin.Context) service.AuthUser {
	user, _ := c.Get(userContextKey)
	return user.(service.AuthUser)
}
