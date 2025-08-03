package main

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rtk-rnjn/ping/routes/internals"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/models"
	"github.com/rtk-rnjn/ping/routes"
)

func getTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.Message{}, &models.UserChannel{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}
	return db
}

func TestHashPasswordAndCheck(t *testing.T) {
	pass := "secret"
	hash, err := internals.HashPassword(pass)
	assert.NoError(t, err)
	assert.True(t, internals.CheckPasswordHash(hash, pass))
	assert.False(t, internals.CheckPasswordHash(hash, "wrong"))
}

func TestGenerateJWT(t *testing.T) {
	user := &models.User{ID: 1, Username: "test"}
	token, err := internals.GenerateJWT(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRegisterUser(t *testing.T) {
	db := getTestDB()
	controller.InitRedis()
	user := &models.User{Username: "testuser", PasswordHash: "hashedpass", DisplayName: "Test User"}
	err := controller.CreateUser(db, user)
	assert.NoError(t, err)
	got, err := controller.GetUserByID(db, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, got.Username)
}

func TestLoginUser(t *testing.T) {
	db := getTestDB()
	controller.InitRedis()

	hashedPass, _ := internals.HashPassword("hashedpass")

	user := &models.User{Username: "testuser", PasswordHash: hashedPass, DisplayName: "Test User"}
	db.Create(user)
	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.PasswordHash)

	token, err := internals.LoginUser(db, user.Username, "hashedpass")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestControllerUserCRUD(t *testing.T) {
	db := getTestDB()
	controller.InitRedis()
	user := &models.User{Username: "u", PasswordHash: "h", DisplayName: "d"}
	err := controller.CreateUser(db, user)
	assert.NoError(t, err)

	got, err := controller.GetUserByID(db, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, got.Username)
}

func TestControllerChannelCRUD(t *testing.T) {
	db := getTestDB()
	controller.InitRedis()
	ch := &models.Channel{Name: "c", Description: "desc"}
	err := controller.CreateChannel(db, ch)
	assert.NoError(t, err)

	got, err := controller.GetChannelByID(db, ch.ID)
	assert.NoError(t, err)
	assert.Equal(t, ch.Name, got.Name)

	ch.Description = "newdesc"
	err = controller.UpdateChannel(db, ch)
	assert.NoError(t, err)

	err = controller.DeleteChannel(db, ch.ID)
	assert.NoError(t, err)
}

func TestControllerMessageCRUD(t *testing.T) {
	db := getTestDB()
	controller.InitRedis()
	user := &models.User{Username: "u", PasswordHash: "h", DisplayName: "d"}
	ch := &models.Channel{Name: "c", Description: "desc"}
	db.Create(user)
	db.Create(ch)
	msg := &models.Message{ChannelID: ch.ID, UserID: user.ID, Content: "hi"}
	err := controller.CreateMessage(db, msg)
	assert.NoError(t, err)

	got, err := controller.GetMessageByID(db, msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, msg.Content, got.Content)
}

func TestControllerUserChannel(t *testing.T) {
	db := getTestDB()
	controller.InitRedis()
	user := &models.User{Username: "u", PasswordHash: "h", DisplayName: "d"}
	ch := &models.Channel{Name: "c", Description: "desc"}
	db.Create(user)
	db.Create(ch)
	uc := &models.UserChannel{UserID: user.ID, ChannelID: ch.ID}
	err := controller.AddUserToChannel(db, uc)
	assert.NoError(t, err)

	list, err := controller.GetUserChannels(db, user.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

	list2, err := controller.GetChannelUsers(db, ch.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, list2)

	err = controller.RemoveUserFromChannel(db, user.ID, ch.ID)
	assert.NoError(t, err)
}

func TestRoutesHandlers(t *testing.T) {
	db := getTestDB()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", routes.RegisterHandler(db))
	r.POST("/login", routes.LoginHandler(db))
	r.GET("/health/database", routes.DatabaseHealthCheck(db))
	r.GET("/health/redis", routes.RedisHealthCheck)
}

func TestControllerCache(t *testing.T) {
	controller.InitRedis()
	key := "testkey"
	val := "testval"
	err := controller.SetCache(key, val, time.Second)
	assert.NoError(t, err)
	got, err := controller.GetCache(key)
	assert.NoError(t, err)
	assert.Equal(t, val, got)
	err = controller.DeleteCache(key)
	assert.NoError(t, err)
}

func TestControllerRunCommand(t *testing.T) {
	controller.InitRedis()
	res, err := controller.RunCommand("PING")
	assert.NoError(t, err)
	assert.Contains(t, res, "PONG")
}
