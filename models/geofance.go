package models

import "encoding/json"

type Geofence struct {
	ID       uint            `json:"id"`
	Name     string          `json:"name"`
	Category string          `json:"category"`
	GeoJSON  json.RawMessage `json:"geojson" gorm:"column:geojson;type:jsonb"`
}
