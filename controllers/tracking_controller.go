package controllers

import (
	"fmt"
	"net/http"
	"time"

	"tracking-api/config"
	"tracking-api/models"
	"tracking-api/package/utils"

	"github.com/gin-gonic/gin"
)

func UpdateDriverLocation(c *gin.Context) {
	var input models.DriverTracking
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB
	now := time.Now()

	var user models.User
	if err := db.First(&user, "id = ?", input.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	input.Name = utils.GetNameFromEmail(user.Email)

	var booking models.Booking
	if err := db.Where(
		"user_id = ? AND is_active = true",
		input.UserID,
	).First(&booking).Error; err != nil {

		db.Model(&models.DriverTracking{}).
			Where("user_id = ? AND is_active = true", input.UserID).
			Updates(map[string]interface{}{
				"is_active":  false,
				"updated_at": now,
			})

		c.JSON(http.StatusOK, gin.H{
			"is_active": false,
			"message":   "No active booking, driver inactive",
		})
		return
	}

	zone, isInside := utils.ValidateZoneFromDatabase(input.Lat, input.Lng)

	if !isInside {
		db.Model(&models.DriverTracking{}).
			Where("user_id = ? AND is_active = true", input.UserID).
			Updates(map[string]interface{}{
				"is_active":  false,
				"updated_at": now,
			})

		c.Status(http.StatusNoContent)
		return
	}

	input.GeofenceName = zone.Name
	input.IsActive = true
	input.BookingID = booking.ID

	switch zone.Category {
	case "port":
		input.Status = "fit"

	case "terminal":
		if zone.Name != booking.TerminalName {
			input.Status = "not match"
		} else {
			input.Status = "fit"
		}

	case "depo":
		input.Status = "other activity"
	}

	var existing models.DriverTracking
	tx := db.Where("user_id = ? AND is_active = true", input.UserID).First(&existing)

	if tx.RowsAffected > 0 {
		distance := utils.CalculateDistance(existing.Lat, existing.Lng, input.Lat, input.Lng)
		if distance > 1 {
			db.Model(&existing).Updates(map[string]interface{}{
				"name":          input.Name,
				"lat":           input.Lat,
				"lng":           input.Lng,
				"geofence_name": input.GeofenceName,
				"status":        input.Status,
				"booking_id":    input.BookingID,
				"is_active":     true,
				"updated_at":    now,
			})
		}
	} else {
		input.CreatedAt = now
		input.UpdatedAt = now
		db.Create(&input)
	}

	broadcastAllActiveDrivers()

	c.JSON(http.StatusOK, gin.H{
		"status":        input.Status,
		"is_active":     true,
		"geofence_name": input.GeofenceName,
		"message":       fmt.Sprintf("Driver %s berada di %s", input.Name, input.GeofenceName),
	})
}

func GetActiveDriverLocations(c *gin.Context) {
	db := config.DB
	var trackings []models.DriverTracking
	db.Find(&trackings)

	c.JSON(http.StatusOK, gin.H{
		"data": trackings,
	})
}
