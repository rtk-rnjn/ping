package routes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rtk-rnjn/ping/config"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{}

type CreateMessageRequest struct {
	ChannelID uint64 `json:"channel_id"`
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
	channelIDUint, err := extractChannelID(c)
	if err != nil {
		return
	}

	if !verifyUserMembership(c, channelIDUint) {
		return
	}

	handleWebSocket(c, channelIDUint)
}

func verifyUserMembership(c *gin.Context, channelIDUint uint64) bool {
	user := c.MustGet("user").(*models.User)
	log.Printf("[INFO] User %d connected to WebSocket for channelID=%d", user.ID, channelIDUint)

	exists, err := controller.IsUserInChannel(config.DB, user.ID, channelIDUint)
	if err != nil {
		log.Printf("[ERROR] Failed to check user channel membership (channelID=%d): %v", channelIDUint, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check channel membership"})
		return false
	}

	if !exists {
		log.Printf("[WARN] User %d not in channel %d, closing WebSocket", user.ID, channelIDUint)
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this channel"})
		return false
	}

	return true
}

func handleWebSocket(c *gin.Context, channelIDUint uint64) {
	cacheKey := fmt.Sprintf("channel:%d:messages", channelIDUint)
	pubSub := controller.Rdb.Subscribe(context.Background(), cacheKey)
	defer closePubSub(pubSub, channelIDUint)

	ch := pubSub.Channel()

	conn, err := upgradeWebSocket(c, channelIDUint)
	if err != nil {
		return
	}
	defer closeWebSocket(conn, channelIDUint)

	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("[INFO] WebSocket closed (channelID=%d): %d %s", channelIDUint, code, text)
		return nil
	})

	go listenToClient(conn, channelIDUint)

	for msg := range ch {
		messageIntUint, err := strconv.ParseUint(msg.Payload, 10, 64)
		if err != nil {
			log.Printf("[ERROR] Invalid message format received on channel %s: %v", cacheKey, err)
			continue
		}

		message, err := controller.GetMessageByID(config.DB, messageIntUint)
		if err != nil {
			log.Printf("[ERROR] Failed to get message by ID %d: %v", messageIntUint, err)
			continue
		}
		if message == nil {
			log.Printf("[WARN] Message with ID %d not found in database", messageIntUint)
			continue
		}
		if message.ChannelID != channelIDUint {
			log.Printf("[WARN] Message with ID %d does not belong to channel %d", messageIntUint, channelIDUint)
			continue
		}
		log.Printf("[DEBUG] Received message from Redis PubSub (channelID=%d): %s", channelIDUint, msg.Payload)

		jsonStringMsg, err := message.ToJSONStringPayload()
		if err != nil {
			log.Printf("[ERROR] Failed to convert message to JSON string (channelID=%d): %v", channelIDUint, err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(jsonStringMsg)); err != nil {
			log.Printf("[ERROR] Failed to send message to WebSocket (channelID=%d): %v", channelIDUint, err)
			break
		}
		log.Printf("[DEBUG] Sent message to client (channelID=%d): %s", channelIDUint, jsonStringMsg)
	}
}

func listenToClient(conn *websocket.Conn, channelIDUint uint64) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[ERROR] Read error from client (channelID=%d): %v", channelIDUint, err)
			break
		}
	}
}

func extractChannelID(c *gin.Context) (uint64, error) {
	channelID := c.Param("channelID")
	if channelID == "" {
		log.Println("[WARN] WebSocket request with missing channelID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel ID is required"})
		return 0, errors.New("missing channelID")
	}

	channelIDUint, err := strconv.ParseUint(channelID, 10, 32)
	if err != nil {
		log.Printf("[ERROR] Invalid channelID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Channel ID"})
		return 0, err
	}

	return channelIDUint, nil
}

func upgradeWebSocket(c *gin.Context, channelID uint64) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to upgrade to WebSocket: %v", err)
		return nil, err
	}

	log.Printf("[INFO] WebSocket connection established for channelID=%d", channelID)
	return conn, nil
}

func closeWebSocket(conn *websocket.Conn, channelID uint64) {
	log.Printf("[INFO] Closing WebSocket connection for channelID=%d", channelID)
	conn.Close()
}

func closePubSub(pubSub *redis.PubSub, channelID uint64) {
	log.Printf("[INFO] Closing Redis PubSub for channelID=%d", channelID)
	pubSub.Close()
}
