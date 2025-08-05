package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{}

type CreateMessageRequest struct {
	ChannelID uint   `json:"channel_id"`
	Content   string `json:"content"`
}

func CreateMessageHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			log.Println("[WARN] Unauthorized access attempt to CreateMessageHandler")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req CreateMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message := &models.Message{
			UserID:    user.(*models.User).ID,
			ChannelID: req.ChannelID,
			Content:   req.Content,
		}

		log.Printf("[INFO] Creating message by userID=%d in channelID=%d", message.UserID, message.ChannelID)
		err := controller.CreateMessage(db, message)
		if err != nil {
			log.Printf("[ERROR] Failed to create message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		log.Printf("[INFO] Message created successfully by userID=%d in channelID=%d", message.UserID, message.ChannelID)
		c.JSON(http.StatusOK, gin.H{"message": "Message created successfully"})
	}
}

func WebSocketChannelMessageHandler(c *gin.Context) {
	channelID := c.Param("channelID")
	if channelID == "" {
		log.Println("[WARN] WebSocket request with missing channelID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel ID is required"})
		return
	}

	channelIDUint, err := strconv.ParseUint(channelID, 10, 32)
	if err != nil {
		log.Printf("[ERROR] Invalid channelID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Channel ID"})
		return
	}

	cacheKey := fmt.Sprintf("channel:%d:messages", channelIDUint)
	pubSub := controller.Rdb.Subscribe(context.Background(), cacheKey)
	defer func() {
		log.Printf("[INFO] Closing Redis PubSub for channelID=%d", channelIDUint)
		pubSub.Close()
	}()

	ch := pubSub.Channel()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer func() {
		log.Printf("[INFO] Closing WebSocket connection for channelID=%d", channelIDUint)
		conn.Close()
	}()

	log.Printf("[INFO] WebSocket connection established for channelID=%d", channelIDUint)

	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("[ERROR] Setting initial read deadline failed: %v", err)
		return
	}

	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("[ERROR] Pong deadline update failed: %v", err)
			return err
		}
		return nil
	})

	go clientPinger(conn, channelIDUint)

	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Printf("[ERROR] Failed to send message to WebSocket (channelID=%d): %v", channelIDUint, err)
			break
		}
		log.Printf("[DEBUG] Sent message to client (channelID=%d): %s", channelIDUint, msg.Payload)
	}
}

func clientPinger(conn *websocket.Conn, channelID uint64) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("[ERROR] Ping failed (channelID=%d): %v", channelID, err)
			return
		}
		log.Printf("[DEBUG] Ping sent to client (channelID=%d)", channelID)
	}
}
