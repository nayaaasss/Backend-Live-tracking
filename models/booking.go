package models

import "time"

type Booking struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id"`
	PortID          int       `json:"port_id"`
	PortName        string    `json:"port_name"`
	TerminalName    string    `json:"terminal_name"`
	GateInPlan      string    `json:"gate_in_plan"`
	ShiftInPlan     string    `json:"shift_in_plan"`
	ContainerNo     string    `json:"container_no"`
	ContainerType   string    `json:"container_type"`
	ContainerSize   string    `json:"container_size"`
	ContainerStatus string    `json:"container_status"`
	ISOCode         string    `json:"iso_code"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
