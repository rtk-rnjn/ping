package internals

import (
	"fmt"
	"log"
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
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", err)
	}
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("[WARN] Password hash mismatch")
	}
	return err == nil
}

func GenerateJWT(user *models.User) (string, error) {
	log.Printf("[INFO] Generating JWT for userID=%d", user.ID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("[ERROR] Failed to sign JWT: %v", err)
	}
	return signedToken, err
}

func ValidateJWT(tokenString string) (*models.User, error) {
	log.Printf("[INFO] Validating JWT: %.10s...", tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := fmt.Sprintf("unexpected signing method: %v", token.Header["alg"])
			log.Printf("[ERROR] %s", errMsg)
			return nil, fmt.Errorf("%s", errMsg)
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		log.Printf("[ERROR] Invalid or expired token: %v", err)
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[ERROR] JWT claims could not be parsed")
		return nil, fmt.Errorf("invalid claims")
	}

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		log.Println("[ERROR] JWT missing or malformed 'id' claim")
		return nil, fmt.Errorf("invalid user ID in token")
	}
	userID := uint(userIDFloat)

	log.Printf("[INFO] Extracted userID=%d from token", userID)
	user, err := controller.GetUserByID(config.DB, userID)
	if err != nil || user == nil {
		log.Printf("[ERROR] User not found for userID=%d: %v", userID, err)
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func RegisterUser(db *gorm.DB, username string, password string, displayName string) (string, error) {
	log.Printf("[INFO] Registering new user: username='%s'", username)

	hashed, err := HashPassword(password)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password for user '%s': %v", username, err)
		return "", err
	}

	user := models.User{
		Username:     username,
		PasswordHash: hashed,
		DisplayName:  displayName,
	}

	if err := controller.CreateUser(config.DB, &user); err != nil {
		log.Printf("[ERROR] Failed to create user '%s': %v", username, err)
		return "", fmt.Errorf("failed to create user")
	}

	log.Printf("[INFO] User '%s' registered successfully with userID=%d", username, user.ID)
	return GenerateJWT(&user)
}

func LoginUser(db *gorm.DB, username, password string) (string, error) {
	log.Printf("[INFO] Attempting login for username='%s'", username)

	var user models.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		log.Printf("[ERROR] Username '%s' not found: %v", username, err)
		return "", fmt.Errorf("invalid credentials")
	}

	if !CheckPasswordHash(user.PasswordHash, password) {
		log.Printf("[WARN] Invalid password attempt for username='%s'", username)
		return "", fmt.Errorf("invalid credentials")
	}

	log.Printf("[INFO] Login successful for username='%s' (userID=%d)", username, user.ID)
	return GenerateJWT(&user)
}
