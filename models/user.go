package models

import (
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"size:32;not null;unique" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:128;not null" json:"password_hash"`
	DisplayName  string    `gorm:"column:display_name;size:64;not null" json:"display_name"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
