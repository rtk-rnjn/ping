package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rtk-rnjn/ping/models"
)

var (
	Rdb       *redis.Client
	ctx       = context.Background()
	redisAddr = os.Getenv("REDIS_ADDR")
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})
	log.Printf("[INFO] Connecting to Redis at %s", redisAddr)
	if _, err := Rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("[FATAL] Failed to connect to Redis: %v", err)
	}
	log.Println("[INFO] Redis connection established successfully")
}

func PingRedis() error {
	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("[ERROR] Redis ping failed: %v", err)
	}
	return err
}

// --- Generic Cache Helpers ---

func setCacheFields(fields map[string]string, timeout time.Duration) error {
	for k, v := range fields {
		if err := Rdb.Set(ctx, k, v, timeout).Err(); err != nil {
			log.Printf("[ERROR] Failed to set Redis key '%s': %v", k, err)
			return fmt.Errorf("failed to set cache for %s: %w", k, err)
		}
		log.Printf("[DEBUG] Redis SET key='%s'", k)
	}
	return nil
}

func getCacheFields(keys []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, k := range keys {
		val, err := Rdb.Get(ctx, k).Result()
		if err != nil {
			log.Printf("[ERROR] Redis GET failed for key '%s': %v", k, err)
			return nil, fmt.Errorf("failed to get cache for %s: %w", k, err)
		}
		results[k] = val
		log.Printf("[DEBUG] Redis GET key='%s' hit", k)
	}
	return results, nil
}

// --- User Cache ---

func SetCacheUser(user models.User) error {
	prefix := fmt.Sprintf("user:%d", user.ID)
	err := setCacheFields(map[string]string{
		prefix + ":username":      user.Username,
		prefix + ":password_hash": user.PasswordHash,
		prefix + ":display_name":  user.DisplayName,
	}, 10*time.Minute)
	if err != nil {
		log.Printf("[ERROR] Failed to cache user ID=%d: %v", user.ID, err)
		return err
	}
	log.Printf("[INFO] User cached: ID=%d", user.ID)
	return nil
}

func GetCacheUser(id uint64) (*models.User, error) {
	prefix := fmt.Sprintf("user:%d", id)
	keys := []string{prefix + ":username", prefix + ":password_hash", prefix + ":display_name"}

	data, err := getCacheFields(keys)
	if err != nil {
		log.Printf("[WARN] Cache miss for user ID=%d: %v", id, err)
		return nil, err
	}

	log.Printf("[INFO] Cache hit for user ID=%d", id)
	return &models.User{
		ID:           id,
		Username:     data[keys[0]],
		PasswordHash: data[keys[1]],
		DisplayName:  data[keys[2]],
	}, nil
}

// --- Channel Cache ---

func SetCacheChannel(channel models.Channel) error {
	prefix := fmt.Sprintf("channel:%d", channel.ID)
	err := setCacheFields(map[string]string{
		prefix + ":name":        channel.Name,
		prefix + ":description": channel.Description,
	}, 10*time.Minute)
	if err != nil {
		log.Printf("[ERROR] Failed to cache channel ID=%d: %v", channel.ID, err)
		return err
	}
	log.Printf("[INFO] Channel cached: ID=%d", channel.ID)
	return nil
}

func GetCacheChannel(id uint64) (*models.Channel, error) {
	prefix := fmt.Sprintf("channel:%d", id)
	keys := []string{prefix + ":name", prefix + ":description"}

	data, err := getCacheFields(keys)
	if err != nil {
		log.Printf("[WARN] Cache miss for channel ID=%d: %v", id, err)
		return nil, err
	}

	log.Printf("[INFO] Cache hit for channel ID=%d", id)
	return &models.Channel{
		ID:          id,
		Name:        data[keys[0]],
		Description: data[keys[1]],
	}, nil
}

func DeleteCacheChannel(id uint64) error {
	prefix := fmt.Sprintf("channel:%d", id)
	err := Rdb.Del(ctx, prefix+":name", prefix+":description").Err()
	if err != nil {
		log.Printf("[ERROR] Failed to delete channel cache for ID=%d: %v", id, err)
		return err
	}
	log.Printf("[INFO] Deleted cache for channel ID=%d", id)
	return nil
}

// --- Message Cache ---

func SetCacheMessage(message models.Message) error {
	prefix := fmt.Sprintf("message:%d", message.ID)
	fields := map[string]string{
		prefix:                 message.Content,
		prefix + ":user_id":    strconv.Itoa(int(message.UserID)),
		prefix + ":channel_id": strconv.Itoa(int(message.ChannelID)),
	}

	if err := setCacheFields(fields, 10*time.Minute); err != nil {
		log.Printf("[ERROR] Failed to cache message ID=%d: %v", message.ID, err)
		return err
	}

	if err := PushMessageToChannel(message.ChannelID, &message); err != nil {
		log.Printf("[ERROR] Failed to push message to queue: %v", err)
		return err
	}
	if err := PublishMessage(message.ChannelID, &message); err != nil {
		log.Printf("[ERROR] Failed to publish message: %v", err)
		return err
	}
	if err := Rdb.LTrim(ctx, fmt.Sprintf("channel:%d:messages", message.ChannelID), -128, -1).Err(); err != nil {
		log.Printf("[ERROR] Failed to trim message list for channel %d: %v", message.ChannelID, err)
		return fmt.Errorf("failed to trim channel messages cache: %w", err)
	}

	log.Printf("[INFO] Message cached: ID=%d, ChannelID=%d", message.ID, message.ChannelID)
	return nil
}

func GetCacheMessage(id uint64) (*models.Message, error) {
	prefix := fmt.Sprintf("message:%d", id)
	keys := []string{prefix, prefix + ":user_id", prefix + ":channel_id"}

	data, err := getCacheFields(keys)
	if err != nil {
		log.Printf("[WARN] Cache miss for message ID=%d: %v", id, err)
		return nil, err
	}

	userID, _ := strconv.ParseUint(data[keys[1]], 10, 32)
	channelID, _ := strconv.ParseUint(data[keys[2]], 10, 32)

	log.Printf("[INFO] Cache hit for message ID=%d", id)
	return &models.Message{
		ID:        id,
		Content:   data[keys[0]],
		UserID:    uint64(userID),
		ChannelID: uint64(channelID),
	}, nil
}

// --- Channel Message Queue & Pub/Sub ---

func PushMessageToChannel(channelID uint64, message *models.Message) error {
	key := fmt.Sprintf("channel:%d:messages", channelID)
	if err := Rdb.RPush(ctx, key, message.ID).Err(); err != nil {
		log.Printf("[ERROR] Failed to RPush message ID=%d: %v", message.ID, err)
		return fmt.Errorf("failed to push message to channel: %w", err)
	}
	log.Printf("[DEBUG] Message ID=%d pushed to channel queue", message.ID)
	return Rdb.LTrim(ctx, key, -128, -1).Err()
}

func PublishMessage(channelID uint64, message *models.Message) error {
	key := fmt.Sprintf("channel:%d:messages", channelID)
	err := Rdb.Publish(ctx, key, message.ID).Err()
	if err != nil {
		log.Printf("[ERROR] Failed to publish message ID=%d to pubsub: %v", message.ID, err)
	}
	return err
}

func GetMessagesFromChannel(channelID uint64) ([]models.Message, error) {
	key := fmt.Sprintf("channel:%d:messages", channelID)
	ids, err := Rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		log.Printf("[ERROR] Failed to read message queue for channel ID=%d: %v", channelID, err)
		return nil, err
	}

	var messages []models.Message
	for _, idStr := range ids {
		msgID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Printf("[WARN] Skipping invalid message ID '%s': %v", idStr, err)
			continue
		}
		msg, err := GetCacheMessage(uint64(msgID))
		if err != nil {
			log.Printf("[WARN] Could not retrieve cached message ID=%d: %v", msgID, err)
			continue
		}
		messages = append(messages, *msg)
	}
	log.Printf("[INFO] Retrieved %d messages from channel ID=%d", len(messages), channelID)
	return messages, nil
}
