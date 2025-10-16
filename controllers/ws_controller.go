package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"tracking-api/models"
	"tracking-api/utils"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex

// Kirim pesan ke semua client yang terkoneksi
func broadcastLocation(msg models.LocationMessage) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			fmt.Println("WriteJSON error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// WebSocket utama
func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	log.Println("Client connected:", conn.RemoteAddr())

	db, err := sql.Open("postgres", "user=postgres password=kanayariany180208 dbname=tracking sslmode=disable")
	if err != nil {
		log.Println("DB connection error:", err)
		return
	}
	defer db.Close()

	for {
		var msg models.LocationMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("ReadJSON error:", err)
			break
		}

		hasBooking, err := utils.HasActiveBooking(db, msg.UserID, msg.PortID, msg.Slot)
		if err != nil {
			log.Println("Booking check error:", err)
			msg.BookingStatus = "strange"
			msg.ArrivalStatus = "error_checking_arrival"
			broadcastLocation(msg)
			continue
		}

		if !hasBooking {
			msg.BookingStatus = "strange"
			msg.ArrivalStatus = "-"
			broadcastLocation(msg)
			continue
		}

		msg.BookingStatus = "fit"

		geofence, err := utils.GetGeofenceByPortAndSlot(db, msg.PortID, msg.Slot)
		if err != nil {
			msg.ArrivalStatus = "geofence_not_found"
			broadcastLocation(msg)
			continue
		}

		if utils.IsInsideGeofence(msg.Lat, msg.Lng, geofence) {
			arrivalTime := time.Now()
			msg.ArrivalStatus = utils.DetermineArrivalStatus(arrivalTime, geofence.StartTime, geofence.EndTime)
		} else {
			msg.ArrivalStatus = "outside"
		}

		log.Printf("Driver %s -> Booking: %s, Arrival: %s\n", msg.Email, msg.BookingStatus, msg.ArrivalStatus)
		broadcastLocation(msg)
	}

	clientsMutex.Lock()
	delete(clients, conn)
	clientsMutex.Unlock()

	log.Println("Client disconnected:", conn.RemoteAddr())
}
