package models

import (
	"time"
)

type User struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Username    string    `gorm:"size:32;not null;unique" json:"username"`
    PasswordHash string   `gorm:"column:password_hash;size:128;not null" json:"password_hash"`
    DisplayName string    `gorm:"column:display_name;size:64;not null" json:"display_name"`
    CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type Message struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`

    ChannelID uint      `gorm:"not null" json:"channel_id"`
    UserID    uint      `gorm:"not null" json:"user_id"`
    Content   string    `gorm:"size:256;not null" json:"content"`
    ReplyToID *uint     `gorm:"column:reply_to" json:"reply_to,omitempty"`

    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

    Channel   Channel   `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE" json:"-"`
    User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
    ReplyTo   *Message  `gorm:"foreignKey:ReplyToID;constraint:OnDelete:SET NULL" json:"-"`
}

type Channel struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"size:64;not null;unique" json:"name"`
    Description string    `gorm:"type:text" json:"description"`
    CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

    Messages    []Message `gorm:"foreignKey:ChannelID" json:"messages,omitempty"`
}

type UserChannel struct {
    UserID    uint      `gorm:"primaryKey" json:"user_id"`
    ChannelID uint      `gorm:"primaryKey" json:"channel_id"`
    JoinedAt  time.Time `gorm:"autoCreateTime;column:joined_at" json:"joined_at"`

    User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
    Channel Channel `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE" json:"-"`
}