package utils

import (
	"encoding/json"
	"fmt"
	"tracking-api/config"
	"tracking-api/models"
)

type Geofence2 struct {
	ID      uint        `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Lat     float64     `json:"lat,omitempty"`
	Lng     float64     `json:"lng,omitempty"`
	Radius  float64     `json:"radius,omitempty"`
	Polygon [][]float64 `json:"polygon,omitempty"`
}

func EnsurePolygonClosed(ring [][]float64) [][]float64 {
	if len(ring) == 0 {
		return ring
	}
	first := ring[0]
	last := ring[len(ring)-1]
	if first[0] != last[0] || first[1] != last[1] {
		ring = append(ring, first)
	}
	return ring
}

func parseGeoJSON(raw json.RawMessage) ([][]float64, float64, float64, float64) {
	polygon := [][]float64{}
	var lat, lng, radius float64

	fmt.Println(string(raw))

	if len(raw) == 0 || string(raw) == "null" {
		fmt.Println("GeoJSON is empty")
		return polygon, lat, lng, radius
	}

	var geo struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties,omitempty"`
		Geometry   struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geometry"`
	}

	if err := json.Unmarshal(raw, &geo); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return polygon, lat, lng, radius
	}

	fmt.Println("GeoJSON Parsed Type:", geo.Geometry.Type)

	if geo.Properties != nil {
		if v, ok := geo.Properties["radius"].(float64); ok {
			radius = v
		}
	}

	switch geo.Geometry.Type {

	case "Point":
		var coords []float64
		if err := json.Unmarshal(geo.Geometry.Coordinates, &coords); err == nil && len(coords) >= 2 {
			lng = coords[0]
			lat = coords[1]
		}

	case "Polygon":
		var coords [][][]float64
		if err := json.Unmarshal(geo.Geometry.Coordinates, &coords); err == nil && len(coords) > 0 {
			ring := EnsurePolygonClosed(coords[0])

			for _, p := range ring {
				polygon = append(polygon, []float64{p[0], p[1]})
			}

			if len(ring) > 0 {
				lng = ring[0][0]
				lat = ring[0][1]
			}
		}
	}
	return polygon, lat, lng, radius
}

func GetAllGeofences2() ([]Geofence2, error) {
	var rows []models.Geofence
	if err := config.DB.Find(&rows).Error; err != nil {
		return nil, err

	}

	var result []Geofence2

	for _, r := range rows {

		polygon, lat, lng, radius := parseGeoJSON(r.GeoJSON)

		result = append(result, Geofence2{
			ID:      r.ID,
			Name:    r.Name,
			Type:    r.Category,
			Lat:     lat,
			Lng:     lng,
			Radius:  radius,
			Polygon: polygon,
		})
	}

	return result, nil
}
