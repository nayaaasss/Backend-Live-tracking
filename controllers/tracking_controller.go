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
		var existing models.DriverTracking
		tx := db.Where("user_id = ? AND is_active = TRUE", input.UserID).First(&existing)

		if tx.RowsAffected > 0 {

			db.Model(&existing).Updates(map[string]interface{}{
				"is_active":     false,
				"gate_out_time": now,
				"updated_at":    now,
			})

			if existing.BookingID != 0 {
				db.Model(&models.Booking{}).
					Where("id = ? AND is_active = true", existing.BookingID).
					Update("is_active", false)
			}
		}
		c.Status(http.StatusNoContent)
		return

	} else {

		input.GeofenceName = zone.Name
		input.IsActive = true

		var booking models.Booking
		hasBooking := db.Where(
			"user_id = ? AND is_active = true",
			input.UserID,
		).Order("created_at DESC").
			First(&booking).Error == nil

		switch zone.Category {
		case "port":
			if !hasBooking {
				input.Status = "strange"
				input.ArrivalStatus = "unknown"
			} else {
				input.Status = "fit"
				input.BookingID = booking.ID
				input.PortName = booking.PortName
				input.TerminalName = booking.TerminalName
				input.ContainerNo = booking.ContainerNo
				input.ISOCode = booking.ISOCode
				input.ContainerStatus = booking.ContainerStatus
				input.StartTime = booking.StartTime
			}

		case "terminal":
			if !hasBooking {
				input.Status = "strange"

			} else if zone.Name != booking.TerminalName {
				input.Status = "not match"

			} else {
				input.Status = "fit"
			}

		case "depo":
			if !hasBooking {
				input.Status = "strange"
			} else {
				input.Status = "other activity"
			}
		}

		var existing models.DriverTracking
		tx := db.Where("user_id = ? AND is_active = true", input.UserID).First(&existing)

		if tx.RowsAffected > 0 {

			if existing.GateInTime.IsZero() {
				input.GateInTime = now
			} else {
				input.GateInTime = existing.GateInTime
			}

			if hasBooking {
				input.ArrivalStatus = utils.GetArrivalStatusByGateIn(
					input.GateInTime,
					booking.StartTime,
					booking.EndTime,
				)
			} else {
				input.ArrivalStatus = "unknown"
			}

			distance := utils.CalculateDistance(existing.Lat, existing.Lng, input.Lat, input.Lng)

			if distance > 1 {
				db.Model(&existing).Updates(map[string]interface{}{
					"name":             input.Name,
					"lat":              input.Lat,
					"lng":              input.Lng,
					"geofence_name":    input.GeofenceName,
					"status":           input.Status,
					"is_active":        true,
					"arrival_status":   input.ArrivalStatus,
					"terminal_name":    input.TerminalName,
					"port_name":        input.PortName,
					"container_no":     input.ContainerNo,
					"iso_code":         input.ISOCode,
					"container_status": input.ContainerStatus,
					"shift_in_plan":    input.ShiftInPlan,
					"gate_in_time":     input.GateInTime,
					"booking_id":       input.BookingID,
					"updated_at":       now,
				})
			}

		} else {

			input.GateInTime = now
			input.CreatedAt = now
			input.UpdatedAt = now

			if hasBooking {
				input.ArrivalStatus = utils.GetArrivalStatusByGateIn(
					input.GateInTime,
					booking.StartTime,
					booking.EndTime,
				)
			} else {
				input.ArrivalStatus = "unknown"
			}

			db.Create(&input)
		}

		broadcastAllActiveDrivers()

		c.JSON(http.StatusOK, gin.H{
			"status":           input.Status,
			"is_active":        input.IsActive,
			"arrival_status":   input.ArrivalStatus,
			"start_time":       booking.StartTime,
			"gate_in_time":     input.GateInTime,
			"terminal_name":    input.TerminalName,
			"container_no":     input.ContainerNo,
			"container_status": input.ContainerStatus,
			"iso_code":         input.ISOCode,
			"message":          fmt.Sprintf("Driver %s berada di %s", input.Name, input.GeofenceName),
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
