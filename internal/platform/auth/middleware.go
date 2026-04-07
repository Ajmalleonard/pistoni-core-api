package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const userContextKey = "currentUser"

func AuthRequired(jwtManager *JWTManager, lookup func(userID string) (any, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtManager.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized"})
			return
		}
		user, err := lookup(claims.Sub)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized"})
			return
		}
		c.Set(userContextKey, user)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		v, ok := c.Get(userContextKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"statusCode": http.StatusUnauthorized, "message": "Unauthorized"})
			return
		}
		userMap, ok := v.(interface{ GetRole() string })
		if !ok || !allowed[userMap.GetRole()] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"statusCode": http.StatusForbidden, "message": "Forbidden resource"})
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (any, bool) {
	v, ok := c.Get(userContextKey)
	return v, ok
}
