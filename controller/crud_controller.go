package controller

import (
	"fmt"

	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, user *models.User) error {
	err := db.Create(user).Error
	if err == nil {
		if err := SetCacheUser(*user); err != nil {
			return err
		}
	}
	return err
}

func GetUserByID(db *gorm.DB, id uint) (*models.User, error) {
	if user, err := GetCacheUser(id); err == nil {
		return user, nil
	}

	var user models.User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	if err := SetCacheUser(user); err != nil {
		return nil, err
	}

	return &user, nil
}

func CreateChannel(db *gorm.DB, ch *models.Channel) error {
	err := db.Create(ch).Error
	if err == nil {
		if err := SetCacheChannel(*ch); err != nil {
			return err
		}
	}
	return err
}

func GetChannelByID(db *gorm.DB, id uint) (*models.Channel, error) {
	if ch, err := GetCacheChannel(id); err == nil {
		return ch, nil
	}

	var ch models.Channel
	err := db.Preload("Messages").First(&ch, id).Error
	if err != nil {
		return nil, err
	}

	if err := SetCacheChannel(ch); err != nil {
		return nil, err
	}

	return &ch, nil
}

func DeleteChannel(db *gorm.DB, id uint) error {
	if err := db.Delete(&models.Channel{}, id).Error; err != nil {
		return err
	}
	if err := DeleteCacheChannel(id); err != nil {
		return err
	}
	return nil
}

func CreateMessage(db *gorm.DB, msg *models.Message) error {
	err := db.Create(msg).Error
	if err == nil {
		if err := SetCacheMessage(*msg); err != nil {
			return err
		}
	}
	return err
}

func GetMessageByID(db *gorm.DB, id uint) (*models.Message, error) {
	if msg, err := GetCacheMessage(id); err == nil {
		return msg, nil
	}

	var msg models.Message
	err := db.Preload("User").First(&msg, id).Error

	if err != nil {
		return nil, err
	}

	if err := SetCacheMessage(msg); err != nil {
		return nil, err
	}
	return &msg, err
}

func GetMessagesByChannelID(db *gorm.DB, channelID uint) ([]models.Message, error) {
	if messages, err := GetMessagesFromChannel(channelID); err == nil {
		return messages, nil
	}
	return nil, fmt.Errorf("failed to get messages from channel %d", channelID)
}

func AddUserToChannel(db *gorm.DB, uc *models.UserChannel) error {
	err := db.Create(uc).Error
	return err
}

func RemoveUserFromChannel(db *gorm.DB, userID uint, channelID uint) error {
	err := db.Delete(&models.UserChannel{}, "user_id = ? AND channel_id = ?", userID, channelID).Error
	return err
}

func GetUserChannels(db *gorm.DB, userID uint) ([]models.UserChannel, error) {
	var list []models.UserChannel
	err := db.Where("user_id = ?", userID).Preload("Channel").Find(&list).Error
	return list, err
}

func GetChannelUsers(db *gorm.DB, channelID uint) ([]models.UserChannel, error) {
	var list []models.UserChannel
	err := db.Where("channel_id = ?", channelID).Preload("User").Find(&list).Error
	return list, err
}
