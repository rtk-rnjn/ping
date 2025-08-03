package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/rtk-rnjn/ping/models"
	"github.com/redis/go-redis/v9"
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
	if _, err := Rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func PingRedis() error {
	_, err := Rdb.Ping(ctx).Result()
	return err
}

// --- Generic Cache Helpers ---

func setCacheFields(fields map[string]string, timeout time.Duration) error {
	for k, v := range fields {
		if err := Rdb.Set(ctx, k, v, timeout).Err(); err != nil {
			return fmt.Errorf("failed to set cache for %s: %w", k, err)
		}
	}
	return nil
}

func getCacheFields(keys []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, k := range keys {
		val, err := Rdb.Get(ctx, k).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get cache for %s: %w", k, err)
		}
		results[k] = val
	}
	return results, nil
}

// --- User Cache ---

func SetCacheUser(user models.User) error {
	prefix := fmt.Sprintf("user:%d", user.ID)
	return setCacheFields(map[string]string{
		prefix + ":username":      user.Username,
		prefix + ":password_hash": user.PasswordHash,
		prefix + ":display_name":  user.DisplayName,
	}, 10*time.Minute)
}

func GetCacheUser(id uint) (*models.User, error) {
	prefix := fmt.Sprintf("user:%d", id)
	keys := []string{prefix + ":username", prefix + ":password_hash", prefix + ":display_name"}

	data, err := getCacheFields(keys)
	if err != nil {
		return nil, err
	}

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
	return setCacheFields(map[string]string{
		prefix + ":name":        channel.Name,
		prefix + ":description": channel.Description,
	}, 10*time.Minute)
}

func GetCacheChannel(id uint) (*models.Channel, error) {
	prefix := fmt.Sprintf("channel:%d", id)
	keys := []string{prefix + ":name", prefix + ":description"}

	data, err := getCacheFields(keys)
	if err != nil {
		return nil, err
	}

	return &models.Channel{
		ID:          id,
		Name:        data[keys[0]],
		Description: data[keys[1]],
	}, nil
}

func DeleteCacheChannel(id uint) error {
	prefix := fmt.Sprintf("channel:%d", id)
	return Rdb.Del(ctx, prefix+":name", prefix+":description").Err()
}

// --- Message Cache ---

func SetCacheMessage(message models.Message) error {
	prefix := fmt.Sprintf("message:%d", message.ID)
	fields := map[string]string{
		prefix:              message.Content,
		prefix + ":user_id": strconv.Itoa(int(message.UserID)),
		prefix + ":channel_id": strconv.Itoa(int(message.ChannelID)),
	}

	if err := setCacheFields(fields, 10*time.Minute); err != nil {
		return err
	}

	if err := Rdb.Del(ctx, prefix+":reply_to_id").Err(); err != nil {
		return fmt.Errorf("failed to delete message reply_to_id: %w", err)
	}

	if err := PushMessageToChannel(message.ChannelID, &message); err != nil {
		return err
	}
	if err := PublishMessage(message.ChannelID, &message); err != nil {
		return err
	}
	if err := Rdb.LTrim(ctx, fmt.Sprintf("channel:%d:messages", message.ChannelID), -128, -1).Err(); err != nil {
		return fmt.Errorf("failed to trim channel messages cache: %w", err)
	}
	log.Printf("Message %d cached successfully", message.ID)
	return nil
}

func GetCacheMessage(id uint) (*models.Message, error) {
	prefix := fmt.Sprintf("message:%d", id)
	keys := []string{prefix, prefix + ":user_id", prefix + ":channel_id"}

	data, err := getCacheFields(keys)
	if err != nil {
		return nil, err
	}

	userID, _ := strconv.ParseUint(data[keys[1]], 10, 32)
	channelID, _ := strconv.ParseUint(data[keys[2]], 10, 32)

	return &models.Message{
		ID:        id,
		Content:   data[keys[0]],
		UserID:    uint(userID),
		ChannelID: uint(channelID),
	}, nil
}

// --- Channel Message Queue & Pub/Sub ---

func PushMessageToChannel(channelID uint, message *models.Message) error {
	key := fmt.Sprintf("channel:%d:messages", channelID)
	if err := Rdb.RPush(ctx, key, message.ID).Err(); err != nil {
		return fmt.Errorf("failed to push message to channel: %w", err)
	}
	return Rdb.LTrim(ctx, key, -128, -1).Err()
}

func PublishMessage(channelID uint, message *models.Message) error {
	return Rdb.Publish(ctx, fmt.Sprintf("channel:%d:messages", channelID), message.ID).Err()
}

func GetMessagesFromChannel(channelID uint) ([]models.Message, error) {
	key := fmt.Sprintf("channel:%d:messages", channelID)
	ids, err := Rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	for _, idStr := range ids {
		msgID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Printf("Invalid message ID %s: %v", idStr, err)
			continue
		}
		msg, err := GetCacheMessage(uint(msgID))
		if err != nil {
			log.Printf("Could not retrieve message %d: %v", msgID, err)
			continue
		}
		messages = append(messages, *msg)
	}
	return messages, nil
}
