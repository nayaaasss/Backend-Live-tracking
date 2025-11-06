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
	dsn := "host=152.69.222.199 user=postgres password=postgres@dmdc dbname=geofencing port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return nil, err
	}
	err = db.Exec("SET search_path TO geofencing, public;").Error
	if err != nil {
		log.Fatal("Gagal mengatur search_path:", err)
	}
	db.AutoMigrate(&models.User{}, &models.Driver{}, &models.Booking{})

	DB = db
	fmt.Println("Database connected successfully")
	return db, nil
}
