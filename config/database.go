package config

import (
	"fmt"
	"log"
	"tracking-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 1. Definisikan variabel GLOBAL untuk menyimpan DSN
const databaseDSN = "host=152.69.222.199 user=postgres password=postgres@dmdc dbname=postgres port=5432 sslmode=disable"

// 2. Fungsi untuk mendapatkan DSN (agar bisa dipanggil dari luar package)
func GetDSN() string {
	return databaseDSN
}

func ConnectDB() (*gorm.DB, error) {
	dsn := "host=152.69.222.199 user=postgres password=postgres@dmdc dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		return nil, err
	}
	err = db.Exec("SET search_path TO geofencing;").Error
	if err != nil {
		log.Fatal("Gagal mengatur search_path:", err)
	}
	db.AutoMigrate(
		&models.User{},
		&models.Booking{},
		&models.DriverTracking{},
	)

	fmt.Println("Using DSN:", dsn)

	DB = db
	fmt.Println("Database connected successfully")
	return db, nil
}
