package main

import (
	"avito/controllers"
	"avito/database"
	"avito/models"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	config, err := godotenv.Read(".env")

	if err != nil {
		log.Fatal("Error reading .env file")
	}

	err = database.InitDatabase(config)
	if err != nil {
		log.Fatalln("could not create database", err)
	}

	err = database.InitCache(config)
	if err != nil {
		log.Fatalln("could not create cache db", err)
	}

	database.GlobalDB.AutoMigrate(&models.User{})
	database.GlobalDB.AutoMigrate(&models.Banner{}, &models.Tag{})

	r := setupRouter()
	r.Run(":8008")
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/login", controllers.Login)
	r.POST("/signup", controllers.Signup)
	r.GET("/user_banner", controllers.UserBanner)
	r.GET("/banner", controllers.GetBanners)
	r.POST("/banner", controllers.CreateBanner)
	r.PATCH("/banner/:id", controllers.UpdateBannerById)
	r.DELETE("/banner/:id", controllers.DeleteBannerById)

	return r
}
