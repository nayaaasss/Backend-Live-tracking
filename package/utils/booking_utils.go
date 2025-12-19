package utils

import "tracking-api/models"

func ApplyBookingData(input *models.DriverTracking, booking models.Booking) {
	input.BookingID = booking.ID
	input.PortName = booking.PortName
	input.TerminalName = booking.TerminalName
	input.ContainerNo = booking.ContainerNo
	input.ISOCode = booking.ISOCode
	input.ContainerStatus = booking.ContainerStatus
	input.StartTime = booking.StartTime
}
