package controller

import (
	"fmt"
	"log"

	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, user *models.User) error {
	if err := db.Create(user).Error; err != nil {
		log.Printf("[ERROR] DB failed to create user: %v", err)
		return err
	}
	log.Printf("[INFO] Created user in DB: ID=%d", user.ID)

	if err := SetCacheUser(*user); err != nil {
		log.Printf("[WARN] Cache set failed for user ID=%d: %v", user.ID, err)
		return err
	}
	log.Printf("[INFO] Cached user: ID=%d", user.ID)
	return nil
}

func GetUserByID(db *gorm.DB, id uint) (*models.User, error) {
	if user, err := GetCacheUser(id); err == nil {
		log.Printf("[INFO] Cache hit for user ID=%d", id)
		return user, nil
	}
	log.Printf("[WARN] Cache miss for user ID=%d, querying DB", id)

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		log.Printf("[ERROR] Failed to find user ID=%d: %v", id, err)
		return nil, err
	}

	if err := SetCacheUser(user); err != nil {
		log.Printf("[WARN] Failed to cache user ID=%d: %v", id, err)
	}
	return &user, nil
}

func CreateChannel(db *gorm.DB, ch *models.Channel) error {
	if err := db.Create(ch).Error; err != nil {
		log.Printf("[ERROR] Failed to create channel: %v", err)
		return err
	}
	log.Printf("[INFO] Created channel in DB: ID=%d", ch.ID)

	if err := SetCacheChannel(*ch); err != nil {
		log.Printf("[WARN] Failed to cache channel ID=%d: %v", ch.ID, err)
		return err
	}
	log.Printf("[INFO] Cached channel: ID=%d", ch.ID)
	return nil
}

func GetChannelByID(db *gorm.DB, id uint) (*models.Channel, error) {
	if ch, err := GetCacheChannel(id); err == nil {
		log.Printf("[INFO] Cache hit for channel ID=%d", id)
		return ch, nil
	}
	log.Printf("[WARN] Cache miss for channel ID=%d, querying DB", id)

	var ch models.Channel
	if err := db.Preload("Messages").First(&ch, id).Error; err != nil {
		log.Printf("[ERROR] Failed to get channel ID=%d: %v", id, err)
		return nil, err
	}

	if err := SetCacheChannel(ch); err != nil {
		log.Printf("[WARN] Failed to cache channel ID=%d: %v", id, err)
	}
	return &ch, nil
}

func DeleteChannel(db *gorm.DB, id uint) error {
	if err := db.Delete(&models.Channel{}, id).Error; err != nil {
		log.Printf("[ERROR] Failed to delete channel ID=%d: %v", id, err)
		return err
	}
	log.Printf("[INFO] Deleted channel from DB: ID=%d", id)

	if err := DeleteCacheChannel(id); err != nil {
		log.Printf("[WARN] Failed to delete channel cache ID=%d: %v", id, err)
		return err
	}
	log.Printf("[INFO] Deleted channel cache: ID=%d", id)
	return nil
}

func CreateMessage(db *gorm.DB, msg *models.Message) error {
	if err := db.Create(msg).Error; err != nil {
		log.Printf("[ERROR] Failed to create message: %v", err)
		return err
	}
	log.Printf("[INFO] Created message: ID=%d", msg.ID)

	if err := SetCacheMessage(*msg); err != nil {
		log.Printf("[WARN] Failed to cache message ID=%d: %v", msg.ID, err)
		return err
	}
	return nil
}

func GetMessageByID(db *gorm.DB, id uint) (*models.Message, error) {
	if msg, err := GetCacheMessage(id); err == nil {
		log.Printf("[INFO] Cache hit for message ID=%d", id)
		return msg, nil
	}
	log.Printf("[WARN] Cache miss for message ID=%d, querying DB", id)

	var msg models.Message
	if err := db.Preload("User").First(&msg, id).Error; err != nil {
		log.Printf("[ERROR] Failed to get message ID=%d: %v", id, err)
		return nil, err
	}

	if err := SetCacheMessage(msg); err != nil {
		log.Printf("[WARN] Failed to cache message ID=%d: %v", id, err)
	}
	return &msg, nil
}

func GetMessagesByChannelID(db *gorm.DB, channelID uint) ([]models.Message, error) {
	if messages, err := GetMessagesFromChannel(channelID); err == nil {
		log.Printf("[INFO] Got %d messages from channel cache ID=%d", len(messages), channelID)
		return messages, nil
	}
	log.Printf("[ERROR] Failed to get messages from Redis cache for channel ID=%d", channelID)
	return nil, fmt.Errorf("failed to get messages from channel %d", channelID)
}

func AddUserToChannel(db *gorm.DB, uc *models.UserChannel) error {
	if err := db.Create(uc).Error; err != nil {
		log.Printf("[ERROR] Failed to add user to channel: %v", err)
		return err
	}
	log.Printf("[INFO] User %d joined channel %d", uc.UserID, uc.ChannelID)
	return nil
}

func RemoveUserFromChannel(db *gorm.DB, userID uint, channelID uint) error {
	if err := db.Delete(&models.UserChannel{}, "user_id = ? AND channel_id = ?", userID, channelID).Error; err != nil {
		log.Printf("[ERROR] Failed to remove user %d from channel %d: %v", userID, channelID, err)
		return err
	}
	log.Printf("[INFO] User %d left channel %d", userID, channelID)
	return nil
}

func GetUserChannels(db *gorm.DB, userID uint) ([]models.UserChannel, error) {
	var list []models.UserChannel
	err := db.Where("user_id = ?", userID).Preload("Channel").Find(&list).Error
	if err != nil {
		log.Printf("[ERROR] Failed to fetch channels for user ID=%d: %v", userID, err)
	}
	return list, err
}

func GetChannelUsers(db *gorm.DB, channelID uint) ([]models.UserChannel, error) {
	var list []models.UserChannel
	err := db.Where("channel_id = ?", channelID).Preload("User").Find(&list).Error
	if err != nil {
		log.Printf("[ERROR] Failed to fetch users for channel ID=%d: %v", channelID, err)
	}
	return list, err
}
