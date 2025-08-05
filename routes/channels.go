package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)

func JoinChannelHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			log.Println("[WARN] Unauthorized attempt to join channel")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			ChannelID uint `json:"channel_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] JoinChannelHandler: Invalid JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[INFO] UserID=%d attempting to join ChannelID=%d", user.(*models.User).ID, req.ChannelID)

		err := controller.AddUserToChannel(db, &models.UserChannel{
			UserID:    user.(*models.User).ID,
			ChannelID: req.ChannelID,
		})
		if err != nil {
			log.Printf("[ERROR] Failed to join channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join channel"})
			return
		}

		log.Printf("[INFO] UserID=%d joined ChannelID=%d successfully", user.(*models.User).ID, req.ChannelID)
		c.JSON(http.StatusOK, gin.H{"message": "Joined channel successfully"})
	}
}

func LeaveChannelHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			log.Println("[WARN] Unauthorized attempt to leave channel")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			ChannelID uint `json:"channel_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] LeaveChannelHandler: Invalid JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[INFO] UserID=%d attempting to leave ChannelID=%d", user.(*models.User).ID, req.ChannelID)

		err := controller.RemoveUserFromChannel(db, user.(*models.User).ID, req.ChannelID)
		if err != nil {
			log.Printf("[ERROR] Failed to leave channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave channel"})
			return
		}

		log.Printf("[INFO] UserID=%d left ChannelID=%d successfully", user.(*models.User).ID, req.ChannelID)
		c.JSON(http.StatusOK, gin.H{"message": "Left channel successfully"})
	}
}

func GetChannelUsersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChannelID uint `json:"channel_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] GetChannelUsersHandler: Invalid JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[INFO] Fetching users for ChannelID=%d", req.ChannelID)

		users, err := controller.GetChannelUsers(db, req.ChannelID)
		if err != nil {
			log.Printf("[ERROR] Failed to get users for ChannelID=%d: %v", req.ChannelID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channel users"})
			return
		}

		log.Printf("[INFO] Found %d users in ChannelID=%d", len(users), req.ChannelID)
		c.JSON(http.StatusOK, users)
	}
}

func CreateChannelHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] CreateChannelHandler: Invalid JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, exists := c.Get("user")
		if !exists {
			log.Println("[WARN] Unauthorized attempt to create channel")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		log.Printf("[INFO] UserID=%d creating channel with name='%s'", user.(*models.User).ID, req.Name)

		channel := &models.Channel{Name: req.Name}
		err := controller.CreateChannel(db, channel)
		if err != nil {
			log.Printf("[ERROR] Failed to create channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
			return
		}

		err = controller.AddUserToChannel(db, &models.UserChannel{
			UserID:    user.(*models.User).ID,
			ChannelID: channel.ID,
		})
		if err != nil {
			log.Printf("[ERROR] Failed to add user to new channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join newly created channel"})
			return
		}

		log.Printf("[INFO] Channel created successfully with ID=%d by UserID=%d", channel.ID, user.(*models.User).ID)
		c.JSON(http.StatusOK, channel)
	}
}
