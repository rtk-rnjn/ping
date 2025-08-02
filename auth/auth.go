package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateJWT(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func RegisterUser(db *gorm.DB, username string, password string, displayName string) (string, error) {
	hashed, err := HashPassword(password)
	if err != nil {
		return "", err
	}
	user := models.User{
		Username:     username,
		PasswordHash: hashed,
		DisplayName:  displayName,
	}
	err = db.Create(&user).Error
	if err != nil {
		return "", err
	}
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	s, _ := json.Marshal(user)
	controller.Rdb.Set(context.Background(), cacheKey, s, time.Minute*10)
	return GenerateJWT(&user)
}

func LoginUser(db *gorm.DB, username, password string) (string, error) {
	var user models.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return "", err
	}
	if !CheckPasswordHash(user.PasswordHash, password) {
		return "", fmt.Errorf("invalid password")
	}
	return GenerateJWT(&user)
}

func ResetToken(db *gorm.DB, userID uint) (string, error) {
	var user models.User
	err := db.First(&user, userID).Error
	if err != nil {
		return "", err
	}
	return GenerateJWT(&user)
}
