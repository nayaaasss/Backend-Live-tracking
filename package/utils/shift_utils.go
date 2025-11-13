package utils

import (
	"time"
)

func GetArrivalStatus(shiftPlan string, t time.Time) string {
	hour := t.Hour()
	minute := t.Minute()
	currentTime := hour*60 + minute

	var shiftStart, shiftEnd int

	switch shiftPlan {
	case "1", "shift 1":
		shiftStart, shiftEnd = 0, 7*60+59
	case "2", "shift 2":
		shiftStart, shiftEnd = 8*60, 15*60+59
	case "3", "shift 3":
		shiftStart, shiftEnd = 16*60, 23*60+59
	default:
		return "unknown"
	}

	if currentTime < shiftStart {
		return "early"
	} else if currentTime <= shiftEnd {
		return "on time"
	} else {
		return "late"
	}
}
