package utils

import (
	"tracking-api/config"
	"tracking-api/models"
)

func ValidateZoneFromDatabase(lat, lng float64) (*models.Zone, bool) {
	var zones []models.Zone
	if err := config.DB.Find(&zones).Error; err != nil {
		return nil, false
	}

	var foundDepo, foundTerminal, foundPort *models.Zone

	for _, z := range zones {

		zoneType, polygon, zLat, zLng, radius :=
			ParseZoneGeoJSON(z.GeoJSON)

		if zoneType == "" {
			continue
		}

		isInside := false

		switch zoneType {
		case "polygon":
			isInside = isPointInsidePolygon(lat, lng, polygon)

		case "circle":
			isInside = CalculateDistance(lat, lng, zLat, zLng) <= radius
		}

		if isInside {
			zCopy := z
			switch z.Category {
			case "depo":
				foundDepo = &zCopy
			case "terminal":
				foundTerminal = &zCopy
			case "port":
				foundPort = &zCopy
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
