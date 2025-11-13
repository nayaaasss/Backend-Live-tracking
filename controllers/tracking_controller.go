package controllers

import (
	"fmt"
	"log"
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

	log.Printf("[%s] %s @%s → Lat: %.6f, Lng: %.6f",
		time.Now().Format("15:04:05"),
		input.UserID,
		input.GeofenceName,
		input.Lat,
		input.Lng,
	)

	db := config.DB
	geofence, isInside := utils.ValidateGeofence(input.Lat, input.Lng)
	now := time.Now()

	if !isInside {
		var existing models.DriverTracking
		tx := db.Where("user_id = ? AND is_active = ?", input.UserID, true).First(&existing)
		if tx.RowsAffected > 0 {
			db.Model(&existing).Updates(map[string]interface{}{
				"is_active":     false,
				"gate_out_time": now,
				"updated_at":    now,
			})
			log.Printf("Driver %s GATE OUT at %s", input.UserID, now.Format("2006-01-02 15:04:05"))
		}
		c.Status(http.StatusNoContent)
		return
	}

	input.GeofenceName = geofence.Name
	input.IsActive = true

	var booking models.Booking
	err := db.Where("id = ?", input.BookingID).First(&booking).Error
	if err != nil {
		input.Status = "strange"
		input.ArrivalStatus = "unknown"
	} else {
		input.Status = "fit"
	}

	var existing models.DriverTracking
	tx := db.Where("user_id = ?", input.UserID).First(&existing)

	if tx.RowsAffected > 0 {
		if existing.GateInTime.IsZero() {
			input.GateInTime = now
			// status kedatangan
			input.ArrivalStatus = utils.GetArrivalStatus(input.ShiftInPlan, now)
			log.Printf("Driver %s GATE IN at %s (Shift %s, %s)",
				input.UserID,
				now.Format("2006-01-02 15:04:05"),
				input.ShiftInPlan,
				input.ArrivalStatus,
			)
		} else {
			input.GateInTime = existing.GateInTime
			input.ArrivalStatus = existing.ArrivalStatus
		}

		// Update posisi hanya jika berpindah lebih dari 1 meter
		distance := utils.CalculateDistance(existing.Lat, existing.Lng, input.Lat, input.Lng)
		if distance > 1 {
			db.Model(&existing).Updates(map[string]interface{}{
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
		// Driver baru melakukan insert
		input.GateInTime = now
		input.ArrivalStatus = utils.GetArrivalStatus(input.ShiftInPlan, now)
		input.CreatedAt = now
		input.UpdatedAt = now
		db.Create(&input)
		log.Printf("Driver %s NEW ENTRY - GATE IN at %s (Shift %s, %s)",
			input.UserID,
			now.Format("2006-01-02 15:04:05"),
			input.ShiftInPlan,
			input.ArrivalStatus,
		)
	}

	broadcastLocation(input)

	c.JSON(http.StatusOK, gin.H{
		"status":         input.Status,
		"is_active":      input.IsActive,
		"arrival_status": input.ArrivalStatus,
		"gate_in_time":   input.GateInTime,
		"gate_out_time":  input.GateOutTime,
		"message":        fmt.Sprintf("Driver %s berada di area %s", input.Name, input.GeofenceName),
	})
}
