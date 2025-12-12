package routes

import (
	"net/http"
	"tracking-api/controllers"
	"tracking-api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// Health Check Endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API is running",
			"status":  "OK",
		})
	})

	r.POST("/register", func(c *gin.Context) {
		controllers.Register(c, db)
	})
	r.POST("/login", func(c *gin.Context) {
		controllers.Login(c, db)
	})
	r.POST("/admin/login", func(c *gin.Context) {
		controllers.LoginAdmin(c, db)
	})

	r.GET("/ws", func(c *gin.Context) {
		controllers.WSHandler(c.Writer, c.Request)
	})

	r.GET("/geofences", func(c *gin.Context) {
		controllers.FetchAllGeofences(c)
	})

	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/token", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")

			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"role":    role,
			})
		})
		auth.POST("/drivers", func(c *gin.Context) {
			controllers.CreateDriver(c, db)
		})
		auth.GET("/drivers", func(c *gin.Context) {
			controllers.GetDrivers(c, db)
		})

		auth.POST("/booking", func(c *gin.Context) {
			controllers.CreateBooking(c)
		})

		auth.GET("/bookings", func(c *gin.Context) {
			controllers.GetUserBookings(c)
		})

		auth.PUT("/booking/:id", func(c *gin.Context) {
			controllers.UpdateBooking(c)
		})

		auth.DELETE("/booking/:id", func(c *gin.Context) {
			controllers.DeleteBooking(c)
		})

		auth.POST("/location/update", func(c *gin.Context) {
			controllers.UpdateDriverLocation(c)
		})

		auth.GET("/location/active", func(c *gin.Context) {
			controllers.GetActiveDriverLocations(c)
		})

	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminOnly())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome Admin"})
		})
	}
}
