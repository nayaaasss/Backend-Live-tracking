package controllers

import (
	"net/http"
	"tracking-api/package/utils"

	"github.com/gin-gonic/gin"
)

func FetchAllGeofences(c *gin.Context) {
	geofences, err := utils.GetAllGeofences2()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": geofences,
	})

}
