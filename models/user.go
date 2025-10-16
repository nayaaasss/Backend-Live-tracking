package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique" json:"email"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	Bookings  []Booking `gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"created_at"`
}
