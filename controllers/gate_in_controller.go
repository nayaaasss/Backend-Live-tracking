package controllers

import (
	"net/http"
	"time"
	"tracking-api/config"
	"tracking-api/models"

	"github.com/gin-gonic/gin"
)

func GateInTime(c *gin.Context) {
	var input struct {
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB
	now := time.Now()

	var tracking models.DriverTracking
	if err := db.Where("user_id = ? AND is_active = true", input.UserID).
		First(&tracking).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active tracking not found"})
		return
	}

	if tracking.GateInTime.IsZero() || tracking.ArrivalStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Arrival status not initialized from geofence",
		})
		return
	}

	db.Model(&tracking).Updates(map[string]interface{}{
		"gate_in_real_time": now,
		"updated_at":        now,
	})

	broadcastAllActiveDrivers()

	c.JSON(http.StatusOK, gin.H{
		"message":           "Gate in confirmed",
		"gate_in_logical":   tracking.GateInTime,
		"gate_in_real_time": now,
		"arrival_status":    tracking.ArrivalStatus,
		"status":            tracking.Status,
		"terminal_name":     tracking.TerminalName,
	})
}
