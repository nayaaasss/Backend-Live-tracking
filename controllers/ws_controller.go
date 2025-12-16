package controllers

import (
	"log"
	"net/http"
	"sync"
	"time"
	"tracking-api/config"
	"tracking-api/models"
	"tracking-api/package/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()
	log.Println("Client connected:", conn.RemoteAddr())

	go func(c *websocket.Conn) {
		defer func() {
			clientsMutex.Lock()
			delete(clients, c)
			clientsMutex.Unlock()
			c.Close()
			log.Println("Client disconnected:", c.RemoteAddr())
		}()

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Println("ReadMessage error (disconnecting):", err)
				break
			}
		}
	}(conn)
}

func broadcastAllActiveDrivers() {
	var activeDrivers []models.DriverTracking

	if err := config.DB.Where("is_active = ?", true).Find(&activeDrivers).Error; err != nil {
		log.Println("Error fetch active drivers:", err)
		return
	}

	var payload []map[string]interface{}

	for _, d := range activeDrivers {
		zone, inside := utils.ValidateZoneFromDatabase(d.Lat, d.Lng)

		item := map[string]interface{}{
			"id":               d.ID,
			"user_id":          d.UserID,
			"name":             d.Name,
			"lat":              d.Lat,
			"lng":              d.Lng,
			"is_active":        d.IsActive,
			"status":           d.Status,
			"arrival_status":   d.ArrivalStatus,
			"terminal_name":    d.TerminalName,
			"port_name":        d.PortName,
			"container_no":     d.ContainerNo,
			"iso_code":         d.ISOCode,
			"container_status": d.ContainerStatus,
			"shift_in_plan":    d.ShiftInPlan,
			"gate_in_time":     d.GateInTime,
			"gate_out_time":    d.GateOutTime,
			"booking_id":       d.BookingID,
			"updated_at":       d.UpdatedAt,
		}

		if inside {
			item["current_geofence"] = zone.Name
			item["geofence_type"] = zone.Category
		} else {
			item["current_geofence"] = nil
			item["geofence_type"] = nil
		}

		payload = append(payload, item)
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		if err := client.WriteJSON(payload); err != nil {
			log.Println("Broadcast error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func StartBroadcastActiveDrivers(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			broadcastAllActiveDrivers()
		}
	}()
}
