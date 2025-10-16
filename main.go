package main

import (
	"log"
	"time"

	"tracking-api/config"
	"tracking-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Koneksi database
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}

	// Router Gin
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register routes
	routes.RegisterRoutes(r, db)

	// Run server
	log.Println("Server running on http://localhost:8080")
	r.Run(":8080")
}
