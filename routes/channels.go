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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			ChannelID uint `json:"channel_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := controller.AddUserToChannel(db, &models.UserChannel{
			UserID:    user.(*models.User).ID,
			ChannelID: req.ChannelID,
		})
		if err != nil {
			log.Printf("Error joining channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Joined channel successfully"})
	}
}

func LeaveChannelHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req struct {
			ChannelID uint `json:"channel_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := controller.RemoveUserFromChannel(db, user.(*models.User).ID, req.ChannelID)
		if err != nil {
			log.Printf("Error leaving channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Left channel successfully"})
	}
}

func GetChannelUsersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChannelID uint `json:"channel_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users, err := controller.GetChannelUsers(db, req.ChannelID)
		if err != nil {
			log.Printf("Error getting channel users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func CreateChannelHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		channel := &models.Channel{Name: req.Name}
		err := controller.CreateChannel(db, channel)
		if err != nil {
			log.Printf("Error creating channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = controller.AddUserToChannel(db, &models.UserChannel{
			UserID:    user.(*models.User).ID,
			ChannelID: channel.ID,
		})
		if err != nil {
			log.Printf("Error adding user to channel: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, channel)
	}
}
