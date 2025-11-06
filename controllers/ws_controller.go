package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"tracking-api/models"
	"tracking-api/package/utils"

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

	db, err := sql.Open("postgres", "host=152.69.222.199 user=postgres password=postgres@dmdc dbname=postgres port=5432 sslmode=disable")
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

		allGeofences, err := utils.GetAllGeofences(db)
		if err != nil {
			log.Println("Gagal ambil geofences dari DB:", err)
			msg.Alert = "Tidak bisa memuat data geofence"
			broadcastLocation(msg)
			continue
		}

		var portGeofence *utils.Geofence
		for _, g := range allGeofences {
			if g.Name == "Pelabuhan Tanjung Priok" {
				portGeofence = g
				break
			}
		}

		// inPelabuhan := false
		// if portGeofence != nil {
		// 	inPelabuhan = utils.IsInsideGeofence(msg.Lat, msg.Lng, portGeofence)
		// }

		// if !inPelabuhan {
		// 	statesMutex.Lock()
		// 	states[msg.UserID] = &driverState{}
		// 	statesMutex.Unlock()

		// 	msg.Alert = "Driver berada di luar area pelabuhan"
		// 	msg.BookingStatus = ""
		// 	msg.ArrivalStatus = "outside_port"
		// 	msg.EnteredAt = time.Time{}
		// 	broadcastLocation(msg)
		// 	continue
		// }

		statesMutex.Lock()
		ds.InPort = true
		statesMutex.Unlock()

		var currentTerminal *utils.Geofence
		for _, g := range allGeofences {
			if g.Name == "Pelabuhan Tanjung Priok" {
				continue
			}
			if utils.IsInsideGeofence(msg.Lat, msg.Lng, g) {
				currentTerminal = g
				break
			}
		}

		if currentTerminal == nil {
			msg.Alert = "Sudah memasuki area pelabuhan"
			msg.BookingStatus = "fit"
			msg.ArrivalStatus = "inside_port"
			msg.EnteredAt = time.Time{}
			broadcastLocation(msg)
			continue
		}

		var booking models.Booking
		err = db.QueryRow(`
			SELECT terminal_name
			FROM bookings
			WHERE user_id = $1 
			ORDER BY created_at DESC
			LIMIT 1`,
			msg.UserID).Scan(&booking.TerminalName)

		if err == sql.ErrNoRows {
			msg.BookingStatus = "Strange"
			msg.ArrivalStatus = "Unknown"
			msg.Destination = "-"
			msg.Alert = fmt.Sprintf("Status: strange (masuk ke %s tanpa booking)", currentTerminal.Name)
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

		msg.Destination = booking.TerminalName

		if booking.TerminalName == currentTerminal.Name {
			msg.BookingStatus = "fit"
			msg.Alert = fmt.Sprintf("Status: fit (Booking dan tujuan sama, yaitu %s)", msg.CurrentTerminal)
		} else {
			msg.BookingStatus = "wrong destination"
			msg.ArrivalStatus = "unknown"
			msg.Alert = fmt.Sprintf("Status: wrong destination (Booking ke %s, tapi masuk %s)", booking.TerminalName, currentTerminal.Name)
			broadcastLocation(msg)
			continue
		}

		msg.Destination = booking.TerminalName
		msg.CurrentTerminal = currentTerminal.Name
		broadcastLocation(msg)

		statesMutex.Lock()
		if ds.FirstTerminal == "" || ds.FirstTerminal != currentTerminal.Name {
			ds.FirstTerminal = currentTerminal.Name
			ds.EnteredAt = time.Now()
			log.Printf("Driver %d masuk ke terminal %s pada %v", msg.UserID, currentTerminal.Name, ds.EnteredAt)
		}
		entered := ds.EnteredAt
		firstTerm := ds.FirstTerminal
		statesMutex.Unlock()

		dbGeofence, err := utils.GetGeofenceByPortAndSlot(db, msg.PortID, msg.Slot)
		if err != nil {
			log.Printf("Tidak bisa menemukan geofence port=%d slot=%s: %v", msg.PortID, msg.Slot, err)
			msg.Alert = "Data geofence untuk slot ini tidak ditemukan"
			msg.ArrivalStatus = "unknown"
			msg.BookingStatus = "strange"
			broadcastLocation(msg)
			continue
		}

		status := utils.DetermineArrivalStatus(entered, dbGeofence.StartTime, dbGeofence.EndTime)

		msg.BookingStatus = "fit"
		msg.ArrivalStatus = status
		msg.EnteredAt = entered

		log.Printf(
			"Driver %d | PortID= %d | Slot= %s | Terminal= %s | Arrival= %v | Start= %v | End= %v | Status= %s",
			msg.UserID, msg.PortID, msg.Slot, firstTerm,
			entered.Format("15:04:05"),
			dbGeofence.StartTime.Format("15:04:05"),
			dbGeofence.EndTime.Format("15:04:05"),
			status,
		)

		if utils.CheckDurationAlert(entered) {
			msg.Alert = fmt.Sprintf("Driver terlalu lama di %s (lebih dari 3 jam)", firstTerm)
		} else {
			switch status {
			case "early":
				msg.Alert = fmt.Sprintf("Status: early (Deteksi pertama di terminal %s)", firstTerm)
			case "late":
				msg.Alert = fmt.Sprintf("Status: late (Deteksi pertama di terminal %s)", firstTerm)
			default:
				msg.Alert = fmt.Sprintf("Status: ontime (Deteksi pertama di terminal %s)", firstTerm)
			}
		}

		broadcastLocation(msg)
	}
}
