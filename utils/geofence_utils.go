package utils

import (
	"database/sql"
	"errors"
	"math"
	"time"
)

type Geofence struct {
	ID        int
	Name      string
	Lat       float64
	Lng       float64
	Radius    float64
	Slot      string
	StartTime time.Time
	EndTime   time.Time
	PortID    int
}

func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const EarthRadius = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}

func GetGeofenceByPortAndSlot(db *sql.DB, portID int, slot string) (*Geofence, error) {
	query := `
		SELECT id, name, lat, lng, radius, slot, start_time, end_time, port_id
		FROM geofances
		WHERE port_id = $1 AND slot = $2
	`
	row := db.QueryRow(query, portID, slot)
	var g Geofence
	err := row.Scan(&g.ID, &g.Name, &g.Lat, &g.Lng, &g.Radius, &g.Slot, &g.StartTime, &g.EndTime, &g.PortID)
	if err != nil {
		return nil, errors.New("geofence not found for port and slot")
	}
	return &g, nil
}

func HasActiveBooking(db *sql.DB, userID, portID int, slot string) (bool, error) {
	query := `
       SELECT id FROM bookings
    	WHERE user_id = $1 AND port_id = $2 AND shift_in_plan = $3
        LIMIT 1
    `
	row := db.QueryRow(query, userID, portID, slot)
	var bookingID int
	err := row.Scan(&bookingID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func IsInsideGeofence(driverLat, driverLng float64, geofence *Geofence) bool {
	return CalculateDistance(driverLat, driverLng, geofence.Lat, geofence.Lng) <= geofence.Radius
}

func DetermineArrivalStatus(arrivalTime, start, end time.Time) string {
	arrival := time.Date(2000, 1, 1, arrivalTime.Hour(), arrivalTime.Minute(), 0, 0, time.UTC)
	startAt := time.Date(2000, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
	endAt := time.Date(2000, 1, 1, end.Hour(), end.Minute(), 0, 0, time.UTC)

	if endAt.Before(startAt) {
		endAt = endAt.Add(24 * time.Hour)
		if arrival.Before(startAt) {
			arrival = arrival.Add(24 * time.Hour)
		}
	}

	if arrival.Before(startAt) {
		return "early"
	} else if arrival.After(endAt) {
		return "late"
	} else {
		return "ontime"
	}
}
