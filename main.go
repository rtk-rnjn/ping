package main

import (
	"log"

	"github.com/rtk-rnjn/ping/routes"
	"github.com/rtk-rnjn/ping/config"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func InitEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}
func InitDB() {
	if err := config.InitDB(); err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	controller.InitRedis()
}

func InitAuthRoutes(r *gin.Engine) *gin.RouterGroup {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	authGroup := r.Group("/auth")
	{

		r.POST("/register", routes.RegisterHandler(config.DB))
		r.POST("/login", routes.LoginHandler(config.DB))
		r.GET("/reset-token", routes.ResetTokenHandler(config.DB))
	}
	return authGroup
}

func InitHealthRoutes(r *gin.Engine) {
	healthGroup := r.Group("/health")
	{
		healthGroup.GET("/database", routes.DatabaseHealthCheck(config.DB))
		healthGroup.GET("/redis", routes.RedisHealthCheck)
	}
}

func main() {
	InitEnv()

	r := gin.Default()
	InitDB()
	InitAuthRoutes(r)
	InitHealthRoutes(r)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	if err := r.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}