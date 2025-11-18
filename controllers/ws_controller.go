package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"
	"tracking-api/config"
	"tracking-api/models"

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

type driverState struct {
	FirstTerminal string
	EnteredAt     time.Time
	InPort        bool
}

var states = make(map[int]*driverState)
var statesMutex sync.Mutex

func broadcastLocation(data interface{}) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		err := client.WriteJSON(data)
		if err != nil {
			log.Println("WriteJSON error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

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

	db, err := config.DB.DB()
	if err != nil {
		log.Println("Gagal ambil koneksi dari GORM:", err)
		return
	}

	for {
		var msg models.LocationMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("ReadJSON error:", err)
			break
		}

		if msg.Slot == "" || msg.PortID == 0 {
			msg.Alert = "Slot atau Port ID belum dikirim dari client"
			msg.ArrivalStatus = "unknown"
			msg.BookingStatus = "strange"
			broadcastLocation(msg)
			continue
		}

		statesMutex.Lock()
		if _, ok := states[msg.UserID]; !ok {
			states[msg.UserID] = &driverState{}
		}
		ds := states[msg.UserID]
		statesMutex.Unlock()

		statesMutex.Lock()
		ds.InPort = true
		statesMutex.Unlock()

		var booking models.Booking
		err = db.QueryRow(`
			SELECT terminal_name
			FROM bookings
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT 1`,
			msg.UserID).Scan(&booking.TerminalName)

		if err == sql.ErrNoRows {
			msg.BookingStatus = "strange"
			msg.ArrivalStatus = "unknown"
			msg.Destination = "-"
			msg.Alert = "Masuk ke terminal tanpa booking"
			broadcastLocation(msg)
			continue
		} else if err != nil {
			log.Println("Error cek Booking:", err)
			msg.BookingStatus = "gagal memeriksa data"
			msg.ArrivalStatus = "gagal memeriksa data"
			msg.Destination = "gagal memeriksa data"
			broadcastLocation(msg)
			continue
		}
	}
}

var mu sync.Mutex

func BroadcastDriverTracking(data interface{}) {
	mu.Lock()
	defer mu.Unlock()

	for conn := range clients {
		conn.WriteJSON(data)
	}
}
