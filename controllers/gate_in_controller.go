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

func GateInTime(c *gin.Context) {
	var input struct {
		UserID     uint      `json:"user_id"`
		BookingID  uint      `json:"booking_id"`
		GateInTime time.Time `json:"gate_in_time"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB
	now := time.Now()

	var booking models.Booking
	if err := db.First(&booking, "id = ?", input.BookingID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking not found"})
		return
	}

	var tracking models.DriverTracking
	if err := db.Where(
		"user_id = ? AND is_active = true",
		input.UserID,
	).First(&tracking).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active tracking not found"})
		return
	}

	if !tracking.GateInTime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gate in already exist"})
		return
	}

	arrivalStatus := utils.GetArrivalStatusByGateIn(
		input.GateInTime,
		booking.StartTime,
		booking.EndTime,
	)

	db.Model(&tracking).Updates(map[string]interface{}{
		"gate_in_time":   input.GateInTime,
		"arrival_status": arrivalStatus,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":        "gate in succes",
		"gate_in_time":   now,
		"arrival_status": arrivalStatus,
	})

	fmt.Println("USER ID:", input.UserID)
	fmt.Println("BOOKING ID:", input.BookingID)
	fmt.Println("GATE IN:", input.GateInTime)

}
