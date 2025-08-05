package models

import (
	"time"
)

type UserChannel struct {
	UserID    uint64    `gorm:"primaryKey" json:"user_id"`
	ChannelID uint64    `gorm:"primaryKey" json:"channel_id"`
	JoinedAt  time.Time `gorm:"autoCreateTime;column:joined_at" json:"joined_at"`

	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Channel Channel `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE" json:"-"`
}
