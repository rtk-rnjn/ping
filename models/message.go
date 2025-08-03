package models

import (
	"time"
)

type Message struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	ChannelID uint   `gorm:"not null" json:"channel_id"`
	UserID    uint   `gorm:"not null" json:"user_id"`
	Content   string `gorm:"size:256;not null" json:"content"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Channel Channel `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE" json:"-"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
