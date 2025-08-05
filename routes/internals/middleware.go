package internals

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func MiddlewareJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("[WARN] Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("[WARN] Malformed Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if len(token) < 10 {
			log.Println("[WARN] Suspiciously short token")
		}

		log.Printf("[INFO] Validating JWT: %.10s...", token) // print first 10 chars for trace

		user, err := ValidateJWT(token)
		if err != nil {
			log.Printf("[ERROR] JWT validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		log.Printf("[INFO] JWT validated successfully for user ID=%d", user.ID)
		c.Set("user", user)
		c.Next()
	}
}
