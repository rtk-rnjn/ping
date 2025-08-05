package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/rtk-rnjn/ping/routes/internals"
)

func MapRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/ping", PingHandler)
	r.GET("/", RootHandler)

	healthGroup := r.Group("/health")
	{
		healthGroup.GET("/database", DatabaseHealthCheck(db))
		healthGroup.GET("/redis", RedisHealthCheck)
	}

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", RegisterHandler(db))
		authGroup.POST("/login", LoginHandler(db))
	}

	channelGroup := r.Group("/channel")
	channelGroup.Use(internals.MiddlewareJWTAuth())
	{
		channelGroup.POST("/join", JoinChannelHandler(db))
		channelGroup.POST("/leave", LeaveChannelHandler(db))
		channelGroup.POST("/users", GetChannelUsersHandler(db))
		channelGroup.POST("/create", CreateChannelHandler(db))
	}

	messageGroup := r.Group("/message")
	messageGroup.Use(internals.MiddlewareJWTAuth())
	{
		messageGroup.POST("/create", CreateMessageHandler(db))
	}

	socketGroup := r.Group("/messages")
	// socketGroup.Use(internals.MiddlewareJWTAuth())
	{
		socketGroup.GET("/:channelID", WebSocketChannelMessageHandler)
	}
}
