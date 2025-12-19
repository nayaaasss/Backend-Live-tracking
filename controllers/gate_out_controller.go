package controllers

import (
	"net/http"
	"time"
	"tracking-api/config"
	"tracking-api/models"

	"github.com/gin-gonic/gin"
)

func GateOut(c *gin.Context) {

	var input struct {
		UserID      uint      `json:"user_id"`
		BookingID   uint      `json:"booking_id"`
		GateOutTime time.Time `json:"gate_out_time"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB
	now := time.Now()

	var tracking models.DriverTracking
	if err := db.Where(
		"user_id = ? AND is_active = true",
		input.UserID,
	).First(&tracking).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active tracking not found"})
		return
	}

	if !tracking.GateOutTime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gate out already exist"})
		return
	}

	db.Model(&tracking).Updates(map[string]interface{}{
		"is_active":     false,
		"gate_out_time": now,
		"updated_at":    now,
	})

	broadcastAllActiveDrivers()

	c.JSON(http.StatusOK, gin.H{
		"message":       "gate out succes",
		"gate_out_time": now,
	})
}
