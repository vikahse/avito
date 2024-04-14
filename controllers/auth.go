package controllers

import (
	"avito/auth"
	"avito/database"
	"avito/models"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Signup(c *gin.Context) {
	var user models.User

	err := c.ShouldBindJSON(&user)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"Error": "Invalid Inputs ",
		})
		c.Abort()
		return
	}

	err = user.HashPassword(user.Password)
	if err != nil {
		log.Println(err.Error())
		c.JSON(500, gin.H{
			"Error": "Error Hashing Password",
		})
		c.Abort()
		return
	}

	err = user.CreateUserRecord()
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"Message": "Sucessfully Register",
	})

}

func Login(c *gin.Context) {
	var payload LoginPayload
	var user models.User

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid Inputs",
		})
		c.Abort()
		return
	}

	result := database.GlobalDB.Where("name = ?", payload.Name).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		c.JSON(401, gin.H{
			"Error": "Invalid User Credentials",
		})
		c.Abort()
		return
	}

	err = user.CheckPassword(payload.Password)
	if err != nil {
		log.Println(err)
		c.JSON(401, gin.H{
			"Error": "Invalid User Credentials",
		})
		c.Abort()
		return
	}

	jwtWrapper := auth.JwtWrapper{
		SecretKey:         "lkamlrmlk12lkm3l2klrfd23",
		Issuer:            "BannerService",
		ExpirationMinutes: 20,
		ExpirationHours:   12,
	}

	signedToken, err := jwtWrapper.GenerateToken(user)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}

	signedtoken, err := jwtWrapper.RefreshToken(user)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}

	tokenResponse := LoginResponse{
		Token:        signedToken,
		RefreshToken: signedtoken,
	}

	c.JSON(200, tokenResponse)
}
