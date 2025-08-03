package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)


type CreateMessageRequest struct {
	ChannelID uint   `json:"channel_id"`
	Content   string `json:"content"`
}

func CreateMessageHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CreateMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message := &models.Message{
			UserID:    user.(*models.User).ID,
			ChannelID: req.ChannelID,
			Content:   req.Content,
		}

		err := controller.CreateMessage(db, message)
		if err != nil {
			log.Printf("Error creating message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Message created successfully"})
	}
}
