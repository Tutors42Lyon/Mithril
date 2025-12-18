package handlers

import (
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/Tutors42Lyon/Mithril/internal/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Session invalide or expire"})
			return
		}

		c.Set("user_id", claims["sub"])
		c.Set("role", claims["role"])

		c.Next()
	}
}