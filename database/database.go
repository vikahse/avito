package database

import (
	"fmt"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"
)

var GlobalDB *gorm.DB

func InitDatabase(config map[string]string) (err error) {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config["DB_HOST"],
		config["DB_USER"],
		config["DB_PASSWORD"],
		config["DB_NAME"],
		config["DB_PORT"],
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return
	}

	GlobalDB = database

	return
}
