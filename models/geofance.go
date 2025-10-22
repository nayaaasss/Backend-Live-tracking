package models

import "time"

type Geofence struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name"`
	Lat       float64    `json:"lat"`
	Lng       float64    `json:"lng"`
	Radius    float64    `json:"radius"`
	PortID    int        `json:"port_id"`
	Slot      string     `json:"slot"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	LatMin    float64    `json:"lat_min"`
	LatMax    float64    `json:"lat_max"`
	LngMin    float64    `json:"lng_min"`
	LngMax    float64    `json:"lng_max"`
}
