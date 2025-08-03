package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/rtk-rnjn/ping/controller"
)

func RedisHealthCheck(c *gin.Context) {
	if err := controller.PingRedis(); err != nil {
		log.Printf("Redis health check failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis is down"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Redis is up"})
}

func DatabaseHealthCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		query := "SELECT 1"
		if err != nil {
			log.Printf("Failed to get database connection: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
			return
		}

		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			log.Printf("Database ping failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is down"})
			return
		}

		if err := sqlDB.QueryRowContext(c.Request.Context(), query).Scan(new(int)); err != nil {
			log.Printf("Database query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			log.Printf("Database ping failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Database is up"})
	}
}
