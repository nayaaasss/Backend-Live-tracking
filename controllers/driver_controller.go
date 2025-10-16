package controllers

import (
	"net/http"
	"tracking-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Create Driver
func CreateDriver(c *gin.Context, db *gorm.DB) {
	var driver models.Driver
	if err := c.ShouldBindJSON(&driver); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// Get All Drivers
func GetDrivers(c *gin.Context, db *gorm.DB) {
	var drivers []models.Driver
	if err := db.Find(&drivers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drivers)
}

// Get Driver by ID
func GetDriverByID(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var driver models.Driver
	if err := db.First(&driver, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}
	c.JSON(http.StatusOK, driver)
}
