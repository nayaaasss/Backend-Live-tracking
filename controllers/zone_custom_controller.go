package controllers

import (
	"net/http"

	"tracking-api/config"
	"tracking-api/models"
	"tracking-api/package/utils"

	"github.com/gin-gonic/gin"
)

type ZoneResponse struct {
	ID       uint        `json:"id"`
	Name     string      `json:"name"`
	Category string      `json:"category"`
	Type     string      `json:"type"`
	Polygon  [][]float64 `json:"polygon,omitempty"`
	Lat      float64     `json:"lat,omitempty"`
	Lng      float64     `json:"lng,omitempty"`
	Radius   float64     `json:"radius,omitempty"`
}

func GetCustomZones(c *gin.Context) {
	var zones []models.Zone

	if err := config.DB.Find(&zones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var result []ZoneResponse

	for _, z := range zones {
		t, poly, lat, lng, radius := utils.ParseZoneGeoJSON(z.GeoJSON)

		if t == "" {
			continue
		}

		result = append(result, ZoneResponse{
			ID:       z.ID,
			Name:     z.Name,
			Category: z.Category,
			Type:     t,
			Polygon:  poly,
			Lat:      lat,
			Lng:      lng,
			Radius:   radius,
		})
	}

	c.JSON(http.StatusOK, result)
}
