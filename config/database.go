package config

import (
	"fmt"
	"log"
	"tracking-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=kanayariany180208 dbname=tracking port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return nil, err
	}

	db.AutoMigrate(&models.User{}, &models.Driver{}, &models.Booking{})

	DB = db
	fmt.Println("Database connected successfully")
	return db, nil
}
