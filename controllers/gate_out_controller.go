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
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.DB

	var tracking models.DriverTracking
	if err := db.Where("user_id = ?", input.UserID).
		Order("created_at DESC").
		First(&tracking).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tracking not found"})
		return
	}

	if tracking.GateOutTime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gate out not detected yet (driver still inside terminal)",
		})
		return
	}

	db.Model(&tracking).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	})

	broadcastAllActiveDrivers()

	c.JSON(http.StatusOK, gin.H{
		"message":       "gate out confirmed",
		"gate_out_time": tracking.GateOutTime,
	})
}
