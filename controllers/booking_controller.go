package controllers

import (
	"net/http"
	"strconv"
	"time"
	"tracking-api/config"
	"tracking-api/models"

	"github.com/gin-gonic/gin"
)

func parseUserID(c *gin.Context) (uint, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return 0, false
	}

	var userID uint
	switch v := userIDInterface.(type) {
	case string:
		id64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot convert user_id"})
			return 0, false
		}
		userID = uint(id64)
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id type"})
		return 0, false
	}

	return userID, true
}

func parseBookingID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return 0, false
	}
	return uint(id), true
}

func CreateBooking(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var book models.Booking
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book.UserID = userID
	book.IsActive = true
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	if err := config.DB.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func GetUserBookings(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	role, _ := c.Get("role")

	if role == "admin" {
		var results []models.Booking
		err := config.DB.Table("bookings").
			Select("bookings.*, users.email as user_email").
			Joins("JOIN users ON users.id = bookings.user_id").
			Scan(&results).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
		return
	}

	// user biasa
	var bookings []models.Booking
	err := config.DB.Where("user_id = ?", userID).Find(&bookings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func UpdateBooking(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	bookingID, ok := parseBookingID(c)
	if !ok {
		return
	}

	var booking models.Booking
	if err := config.DB.Where("id = ? AND user_id = ?", bookingID, userID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	var input models.Booking
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking.PortID = input.PortID
	booking.PortName = input.PortName
	booking.TerminalName = input.TerminalName
	booking.GateInPlan = input.GateInPlan
	booking.ShiftInPlan = input.ShiftInPlan
	booking.ContainerNo = input.ContainerNo
	booking.ContainerType = input.ContainerType
	booking.ContainerSize = input.ContainerSize
	booking.ContainerStatus = input.ContainerStatus
	booking.ISOCode = input.ISOCode
	booking.UpdatedAt = time.Now()
	booking.StartTime = input.StartTime
	booking.EndTime = input.EndTime

	if err := config.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, booking)
}

func DeleteBooking(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	bookingID, ok := parseBookingID(c)
	if !ok {
		return
	}

	var booking models.Booking
	if err := config.DB.Where("id = ? AND user_id = ?", bookingID, userID).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	if err := config.DB.Delete(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking deleted"})
}
