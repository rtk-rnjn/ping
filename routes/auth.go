package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/rtk-rnjn/ping/routes/internals"
)

type AuthRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name,omitempty"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AuthRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] RegisterHandler: Invalid JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[INFO] Attempting to register user: username='%s'", req.Username)

		token, err := internals.RegisterUser(db, req.Username, req.Password, req.DisplayName)
		if err != nil {
			log.Printf("[ERROR] Failed to register user '%s': %v", req.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
			return
		}

		log.Printf("[INFO] User registered successfully: username='%s'", req.Username)
		c.JSON(http.StatusOK, TokenResponse{Token: token})
	}
}

func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AuthRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] LoginHandler: Invalid JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[INFO] Attempting login for user: username='%s'", req.Username)

		token, err := internals.LoginUser(db, req.Username, req.Password)
		if err != nil {
			log.Printf("[WARN] Login failed for user '%s': %v", req.Username, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		log.Printf("[INFO] User logged in successfully: username='%s'", req.Username)
		c.JSON(http.StatusOK, TokenResponse{Token: token})
	}
}
