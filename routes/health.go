package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/rtk-rnjn/ping/controller"
)

func RedisHealthCheck(c *gin.Context) {
	log.Println("[INFO] Performing Redis health check")

	if err := controller.PingRedis(); err != nil {
		log.Printf("[ERROR] Redis health check failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis is down"})
		return
	}

	log.Println("[INFO] Redis is up and responding")
	c.JSON(http.StatusOK, gin.H{"message": "Redis is up"})
}

func DatabaseHealthCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("[INFO] Performing database health check")

		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("[ERROR] Failed to get DB object from GORM: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
			return
		}

		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			log.Printf("[ERROR] Database ping (context) failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is down"})
			return
		}

		query := "SELECT 1"
		if err := sqlDB.QueryRowContext(c.Request.Context(), query).Scan(new(int)); err != nil {
			log.Printf("[ERROR] Database test query failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			log.Printf("[ERROR] Final database ping failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is down"})
			return
		}

		log.Println("[INFO] Database is up and operational")
		c.JSON(http.StatusOK, gin.H{"message": "Database is up"})
	}
}
