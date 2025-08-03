package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rtk-rnjn/ping/config"
	"github.com/rtk-rnjn/ping/controller"
	"github.com/rtk-rnjn/ping/routes"

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

func InitAuthRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	{

		authGroup.POST("/register", routes.RegisterHandler(config.DB))
		authGroup.POST("/login", routes.LoginHandler(config.DB))
	}
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
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	InitDB()
	InitAuthRoutes(r)
	InitHealthRoutes(r)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/", routes.RootHandler)

	if err := r.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
