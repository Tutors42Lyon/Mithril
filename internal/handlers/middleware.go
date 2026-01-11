package handlers

import (
	"log"
	"strings"

	"github.com/Tutors42Lyon/Mithril/internal/utils"
	"github.com/Tutors42Lyon/Mithril/internal/repositories"
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

		c.Next()
	}
}

func IsAdmin(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authentication required"})
			return
		}

		var userID uint
		switch v := userIDValue.(type) {
		case float64:
			userID = uint(v)
		case uint:
			userID = v
		default:
			c.AbortWithStatusJSON(500, gin.H{"error": "Invalid user identification format"})
			return
		}

		user, err := userRepo.GetByID(userID)
		if err != nil || user.Role != "admin" {
			log.Printf("Access denied: User %d is not admin", userID)
			c.AbortWithStatusJSON(403, gin.H{"error": "Access denied: reserved for administrators"})
			return
		}

		c.Next()
	}
}