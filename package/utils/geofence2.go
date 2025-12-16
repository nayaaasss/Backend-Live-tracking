package utils

import (
	"encoding/json"
	"fmt"
	"tracking-api/config"
	"tracking-api/models"
)

type Geofence2 struct {
	ID       uint        `json:"id"`
	Name     string      `json:"name"`
	Category string      `json:"type"`
	Lat      float64     `json:"lat,omitempty"`
	Lng      float64     `json:"lng,omitempty"`
	Radius   float64     `json:"radius,omitempty"`
	Polygon  [][]float64 `json:"polygon,omitempty"`
}

func parseGeoJSONFast(raw json.RawMessage) ([][]float64, float64, float64, float64) {
	polygon := [][]float64{}
	var lat, lng, radius float64

	if len(raw) == 0 || string(raw) == "null" {
		return polygon, lat, lng, radius
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return polygon, lat, lng, radius
	}

	geom, ok := obj["geometry"].(map[string]interface{})
	if !ok {
		return polygon, lat, lng, radius
	}

	geomType, _ := geom["type"].(string)
	coordsRaw := geom["coordinates"]

	switch geomType {
	case "Point":
		if coords, ok := coordsRaw.([]interface{}); ok && len(coords) >= 2 {
			lng, _ = coords[0].(float64)
			lat, _ = coords[1].(float64)
		}

	case "Polygon":
		if coordsArr, ok := coordsRaw.([]interface{}); ok && len(coordsArr) > 0 {
			if ring, ok := coordsArr[0].([]interface{}); ok {
				for _, p := range ring {
					if point, ok := p.([]interface{}); ok && len(point) >= 2 {
						polygon = append(polygon, []float64{point[0].(float64), point[1].(float64)})
					}
				}
				if len(polygon) > 0 {
					first := polygon[0]
					last := polygon[len(polygon)-1]
					if first[0] != last[0] || first[1] != last[1] {
						polygon = append(polygon, first)
					}
					lng = polygon[0][0]
					lat = polygon[0][1]
				}
			}
		}
	}

	if props, ok := obj["properties"].(map[string]interface{}); ok {
		if r, ok := props["radius"].(float64); ok {
			radius = r
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
		polygon, lat, lng, radius := parseGeoJSONFast(r.GeoJSON)

		result = append(result, Geofence2{
			ID:       r.ID,
			Name:     r.Name,
			Category: r.Category,
			Lat:      lat,
			Lng:      lng,
			Radius:   radius,
			Polygon:  polygon,
		})
	}

	return result, nil
}

func isPointInsidePolygon(lat, lng float64, poly [][]float64) bool {
	inside := false
	j := len(poly) - 1

	for i := 0; i < len(poly); i++ {
		xi, yi := poly[i][0], poly[i][1]
		xj, yj := poly[j][0], poly[j][1]

		intersect := ((yi > lat) != (yj > lat)) &&
			(lng < (xj-xi)*(lat-yi)/(yj-yi)+xi)

		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

func ValidateGeofenceFromDatabase(lat, lng float64) (*Geofence2, bool) {
	geofences, err := GetAllGeofences2()
	if err != nil {
		return nil, false
	}

	var foundPort, foundTerminal, foundDepo *Geofence2

	for _, g := range geofences {
		isInside := false

		if g.Category == "terminal" {
			if len(g.Polygon) > 0 {
				isInside = isPointInsidePolygon(lat, lng, g.Polygon)
			} else {
				isInside = false
			}

		} else {
			if len(g.Polygon) > 0 {
				isInside = isPointInsidePolygon(lat, lng, g.Polygon)
			} else if g.Radius > 0 {
				d := CalculateDistance(lat, lng, g.Lat, g.Lng)
				isInside = d <= g.Radius
			}
		}

		if isInside {
			gCopy := g

			switch g.Category {
			case "depo":
				foundDepo = &gCopy
			case "terminal":
				foundTerminal = &gCopy
			case "port":
				foundPort = &gCopy
			}
		}
	}

	if foundDepo != nil {
		return foundDepo, true
	}
	if foundTerminal != nil {
		return foundTerminal, true
	}
	if foundPort != nil {
		return foundPort, true
	}

	return nil, false

}
