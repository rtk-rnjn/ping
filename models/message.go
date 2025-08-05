package models

import (
	"encoding/json"
	"time"
)

type Message struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	ChannelID uint64 `gorm:"not null" json:"channel_id"`
	UserID    uint64 `gorm:"not null" json:"user_id"`
	Content   string `gorm:"size:256;not null" json:"content"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Channel Channel `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE" json:"-"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (m *Message) ToJSONStringPayload() (string, error) {
	payload := map[string]any{
		"id":         m.ID,
		"channel_id": m.ChannelID,
		"user_id":    m.UserID,
		"content":    m.Content,
		"created_at": m.CreatedAt.Format(time.RFC3339),
	}

	jsonString, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}
