package models

import "time"

type DriverTracking struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id"`
	BookingID       uint      `json:"booking_id"`
	Name            string    `json:"name"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	ContainerNo     string    `json:"container_no"`
	ISOCode         string    `json:"iso_code"`
	GeofenceName    string    `json:"geofence_name"`
	TerminalName    string    `json:"terminal_name"`
	GateInTime      time.Time `json:"gate_in_time"`
	GateOutTime     time.Time `json:"gate_out_time"`
	PortName        string    `json:"port_name"`
	ContainerStatus string    `json:"container_status"`
	ShiftInPlan     string    `json:"shift_in_plan"`
	StartTime       time.Time `json:"start_time"`
	Status          string    `json:"status"`
	ArrivalStatus   string    `json:"arrival_status"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (DriverTracking) TableName() string {
	return "geofencing.tracking_driver"
}
