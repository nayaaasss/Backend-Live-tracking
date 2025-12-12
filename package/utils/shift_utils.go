package utils

import "time"

func GetArrivalStatusByGateIn(gateIn, start, end time.Time) string {

	if start.IsZero() || end.IsZero() {
		return "unknown"
	}

	if gateIn.Before(start) {
		return "early"
	}

	if gateIn.After(end) {
		return "late"
	}

	return "ontime"
}
