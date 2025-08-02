package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rtk-rnjn/ping/models"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.Message{},
		&models.UserChannel{},
	)
	if err != nil {
		return err
	}

	return nil
}
