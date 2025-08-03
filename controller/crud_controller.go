package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rtk-rnjn/ping/models"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, user *models.User) error {
	err := db.Create(user).Error
	if err == nil {
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		data, _ := json.Marshal(user)
		Rdb.Set(ctx, cacheKey, data, time.Minute*10)
	}
	return err
}

func GetUserByID(db *gorm.DB, id uint) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)
	if val, err := Rdb.Get(ctx, cacheKey).Result(); err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(val), &user); err != nil {
			return nil, err
		}
		return &user, nil
	}

	var user models.User
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(user)
	Rdb.Set(ctx, cacheKey, data, time.Minute*10)
	return &user, nil
}

func UpdateUser(db *gorm.DB, user *models.User) error {
	err := db.Save(user).Error
	if err == nil {
		cacheKey := fmt.Sprintf("user:%d", user.ID)
		Rdb.Del(ctx, cacheKey)
	}
	return err
}

func DeleteUser(db *gorm.DB, id uint) error {
	err := db.Delete(&models.User{}, id).Error
	if err == nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		Rdb.Del(ctx, cacheKey)
	}
	return err
}

func CreateChannel(db *gorm.DB, ch *models.Channel) error {
	err := db.Create(ch).Error
	if err == nil {
		cacheKey := fmt.Sprintf("channel:%d", ch.ID)
		data, _ := json.Marshal(ch)
		Rdb.Set(ctx, cacheKey, data, time.Minute*10)
	}
	return err
}

func GetChannelByID(db *gorm.DB, id uint) (*models.Channel, error) {
	cacheKey := fmt.Sprintf("channel:%d", id)
	if val, err := Rdb.Get(ctx, cacheKey).Result(); err == nil {
		var ch models.Channel
		if err := json.Unmarshal([]byte(val), &ch); err != nil {
			return nil, err
		}
		return &ch, nil
	}

	var ch models.Channel
	err := db.Preload("Messages").First(&ch, id).Error
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(ch)
	Rdb.Set(ctx, cacheKey, data, time.Minute*10)
	return &ch, nil
}

func UpdateChannel(db *gorm.DB, ch *models.Channel) error {
	err := db.Save(ch).Error
	if err == nil {
		cacheKey := fmt.Sprintf("channel:%d", ch.ID)
		Rdb.Del(ctx, cacheKey)
	}
	return err
}

func DeleteChannel(db *gorm.DB, id uint) error {
	err := db.Delete(&models.Channel{}, id).Error
	if err == nil {
		cacheKey := fmt.Sprintf("channel:%d", id)
		Rdb.Del(ctx, cacheKey)
	}
	return err
}

func CreateMessage(db *gorm.DB, msg *models.Message) error {
	err := db.Create(msg).Error
	if err == nil {
		Rdb.Del(ctx, fmt.Sprintf("channel:%d", msg.ChannelID))
	}
	return err
}

func GetMessageByID(db *gorm.DB, id uint) (*models.Message, error) {
	var msg models.Message
	err := db.Preload("User").Preload("ReplyTo").First(&msg, id).Error
	return &msg, err
}

func GetMessagesByChannelID(db *gorm.DB, channelID uint) ([]models.Message, error) {
	var msgs []models.Message
	err := db.Where("channel_id = ?", channelID).Preload("User").Find(&msgs).Error
	return msgs, err
}

func UpdateMessage(db *gorm.DB, msg *models.Message) error {
	err := db.Save(msg).Error
	if err == nil {
		Rdb.Del(ctx, fmt.Sprintf("channel:%d", msg.ChannelID))
	}
	return err
}

func DeleteMessage(db *gorm.DB, id uint) error {
	var msg models.Message
	err := db.First(&msg, id).Error
	if err != nil {
		return err
	}
	err = db.Delete(&models.Message{}, id).Error
	if err == nil {
		Rdb.Del(ctx, fmt.Sprintf("channel:%d", msg.ChannelID))
	}
	return err
}

func AddUserToChannel(db *gorm.DB, uc *models.UserChannel) error {
	err := db.Create(uc).Error
	if err == nil {
		Rdb.Del(ctx, fmt.Sprintf("user_channels:%d", uc.UserID))
		Rdb.Del(ctx, fmt.Sprintf("channel_users:%d", uc.ChannelID))
	}
	return err
}

func RemoveUserFromChannel(db *gorm.DB, userID uint, channelID uint) error {
	err := db.Delete(&models.UserChannel{}, "user_id = ? AND channel_id = ?", userID, channelID).Error
	if err == nil {
		Rdb.Del(ctx, fmt.Sprintf("user_channels:%d", userID))
		Rdb.Del(ctx, fmt.Sprintf("channel_users:%d", channelID))
	}
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
