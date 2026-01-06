package handlers

import (
	"log"
	"strings"

	"github.com/Tutors42Lyon/Mithril/internal/utils"
	"github.com/gin-gonic/gin"
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

		log.Println(claims)
		c.Next()
	}
}

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")

		log.Println("User role: " , userRole, "exists :", exists)
		if !exists || userRole != "admin" {
			c.JSON(403, gin.H{"error": "Access denied: reserved for administrators"})
			c.Abort()
			return
		}
		c.Next()
	}
}
