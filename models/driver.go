package models

import "time"

type Driver struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	LicenseNo   string    `json:"license_no"`
	Status      string    `json:"status"`
	DriverEmail string    `json:"driver_email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
