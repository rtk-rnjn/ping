package internals

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

func MiddlewareJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		user, err := ValidateJWT(token)
		if err != nil {
			log.Printf("Invalid JWT: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
