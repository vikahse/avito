package models

import (
	"avito/database"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Banner struct {
	gorm.Model
	ID        int                    `gorm:"primary_key" redis:"id"`
	TagIds    []Tag                  `gorm:"many2many:banner_tags" json:"tag_ids" binding:"required" redis:"tag_ids"`
	FeatureId int                    `json:"feature_id" binding:"required" redis:"feature_id"`
	Content   map[string]interface{} `gorm:"serializer:json" json:"content" redis:"content"`
	IsActive  bool                   `json:"is_active" redis:"is_active"`
}

func (b *Banner) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, b)
}

func (b *Banner) MarshalBinary() ([]byte, error) {
	return json.Marshal(b)
}

type Tag struct {
	gorm.Model
	Value int `redis:"value" json:"value"`
}

type User struct {
	gorm.Model
	ID       int    `gorm:"primary_key"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	IsAdmin  bool   `json:"is_admin"`
}

func (banner *Banner) CreateBannerRecord() error {
	result := database.GlobalDB.Create(&banner)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (user *User) CreateUserRecord() error {
	var existingUser User
	result := database.GlobalDB.Where("name = ?", user.Name).First(&existingUser)
	if result.RowsAffected == 0 {
		result := database.GlobalDB.Create(&user)
		if result.Error != nil {
			return result.Error
		}
		return nil
	}
	return errors.New("User already exists")
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) CheckPassword(givenPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(givenPassword))
	if err != nil {
		return err
	}
	return nil
}
