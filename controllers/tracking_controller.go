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

	zone, isInside := utils.ValidateZoneFromDatabase(input.Lat, input.Lng)

	if !isInside {
		db.Model(&models.DriverTracking{}).
			Where("user_id = ? AND is_active = true", input.UserID).
			Updates(map[string]interface{}{
				"is_active":  false,
				"updated_at": now,
			})

		db.Model(&models.Booking{}).
			Where("user_id = ? AND is_active = true", input.UserID).
			Updates(map[string]interface{}{
				"is_active":  false,
				"updated_at": now,
			})

		c.Status(http.StatusNoContent)
		return
	} else {
		input.GeofenceName = zone.Name
		input.IsActive = true

		var booking models.Booking
		hasBooking := db.Where("user_id = ? AND is_active = true", input.UserID).
			Order("created_at DESC").
			First(&booking).Error == nil

		var tracking models.DriverTracking
		db.Where("user_id = ? AND is_active = true", input.UserID).
			First(&tracking)

		switch zone.Category {
		case "port":
			if !hasBooking {
				input.Status = "stranger"
				input.ArrivalStatus = "unknown"
			} else {
				input.Status = "fit"
				utils.ApplyBookingData(&input, booking)

				compareTime := time.Now()

				if !tracking.GateInTime.IsZero() {
					compareTime = tracking.GateInTime
				}

				input.ArrivalStatus = utils.GetArrivalStatusByGateIn(
					compareTime,
					booking.StartTime,
					booking.EndTime,
				)
			}

		case "terminal":
			if !hasBooking {
				input.Status = "stranger"
			} else if zone.Name != booking.TerminalName {
				input.Status = "not match"
			} else {
				input.Status = "fit"
				utils.ApplyBookingData(&input, booking)
			}

		case "depo":
			if !hasBooking {
				input.Status = "stranger"
			} else {
				input.Status = "other activity"
			}

		}

		var existing models.DriverTracking
		tx := db.Where("user_id = ? AND is_active = true", input.UserID).First(&existing)

		if tx.RowsAffected > 0 {
			distance := utils.CalculateDistance(existing.Lat, existing.Lng, input.Lat, input.Lng)
			if distance > 1 {
				db.Model(&existing).Updates(map[string]interface{}{
					"name":           input.Name,
					"lat":            input.Lat,
					"lng":            input.Lng,
					"geofence_name":  input.GeofenceName,
					"status":         input.Status,
					"arrival_status": input.ArrivalStatus,
					"terminal_name":  booking.TerminalName,
					"port_name":      booking.PortName,
					"is_active":      true,
					"updated_at":     now,
				})
			}
		} else {
			input.CreatedAt = now
			input.UpdatedAt = now
			db.Create(&input)
		}

		broadcastAllActiveDrivers()

		c.JSON(http.StatusOK, gin.H{
			"terminal_name":  input.TerminalName,
			"status":         input.Status,
			"arrival_status": input.ArrivalStatus,
			"is_active":      true,
			"geofence_name":  input.GeofenceName,
			"message":        fmt.Sprintf("Driver %s berada di %s", input.Name, input.GeofenceName),
		})
	}

}

func GetActiveDriverLocations(c *gin.Context) {
	db := config.DB
	var trackings []models.DriverTracking
	db.Find(&trackings)

	c.JSON(http.StatusOK, gin.H{
		"data": trackings,
	})

}
