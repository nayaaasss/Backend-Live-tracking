package models

import "time"

type LocationMessage struct {
	UserID          int       `json:"userId"`
	Email           string    `json:"email"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	BookingStatus   string    `json:"booking_status"`
	ArrivalStatus   string    `json:"arrival_status"`
	Slot            string    `json:"slot"`
	PortID          int       `json:"port_id"`
	EnteredAt       time.Time `json:"entered_at"`
	Alert           string    `json:"alert"`
	Destination     string    `json:"destination,omitempty"`
	CurrentTerminal string    `json:"current_terminal"`
}
