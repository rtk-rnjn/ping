package main

import (
	"os"
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


func main() {
	InitEnv()
	InitDB()

	r := gin.Default()

	routes.MapRoutes(r, config.DB)


	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")

	if HOST == "" {
		panic("HOST environment variable is not set")
	}

	if PORT == "" {
		panic("PORT environment variable is not set")
	}

	if err := r.Run(HOST + ":" + PORT); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
