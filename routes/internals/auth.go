package internals

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rtk-rnjn/ping/config"
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

func ValidateJWT(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	userID := uint(claims["id"].(float64))
	user := &models.User{ID: userID}

	if user, err := controller.GetUserByID(config.DB, userID); err != nil {
		return nil, fmt.Errorf("user not found")
	} else if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
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

	if err := controller.CreateUser(config.DB, &user); err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}

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
