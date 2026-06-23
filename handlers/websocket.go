package handlers

import (
	"corehub/database"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]string)
	broadcast = make(chan Message)
	mutex     = &sync.RWMutex{}
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// Message represents the standard structure for all real-time communications.
type Message struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Username         string `json:"username"`
	ReceiverUsername string `json:"receiver_username"`
	GroupID          int    `json:"group_id"`
	Content          string `json:"content"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
}

// HandleConnections upgrades the HTTP connection to a WebSocket and listens for incoming messages.
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value("username")
	if val == nil {
		return
	}
	username := val.(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	mutex.Lock()
	clients[conn] = username
	activeCount := len(clients)
	mutex.Unlock()

	log.Printf("✅ Client connected: [%s] | Total active clients: %d\n", username, activeCount)
	broadcast <- Message{Type: "status", Username: username, Content: "online"}

	defer func() {
		mutex.Lock()
		// حذف اتصال هذه النافذة/الجهاز فقط
		delete(clients, conn)
		
		// 🌟 الضربة الهندسية: فحص ما إذا كان المستخدم لا يزال يمتلك نوافذ/أجهزة أخرى مفتوحة
		stillOnline := false
		for _, uname := range clients {
			if uname == username {
				stillOnline = true
				break
			}
		}

		activeCount = len(clients)
		mutex.Unlock()
		
		conn.Close()

		log.Printf("❌ Client disconnected: [%s] | Total active clients: %d\n", username, activeCount)
		
		// 🌟 لا نرسل حالة "أوفلاين" إلا إذا تم إغلاق جميع نوافذ وأجهزة هذا المستخدم
		if !stillOnline {
			broadcast <- Message{Type: "status", Username: username, Content: "offline"}
		}
	}()

	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(rawMsg, &msg); err == nil {
			msg.Username = username

			if msg.Type == "ping" {
				continue
			}

			if msg.Type == "chat" {
				msg.CreatedAt = time.Now().Format("03:04 PM")
				if msg.GroupID != 0 {
					_, dbErr := database.DB.Exec(
						"INSERT INTO messages (username, content, client_msg_id, status, group_id, created_at) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)",
						msg.Username, msg.Content, msg.ID, "sent", msg.GroupID,
					)
					if dbErr == nil {
						broadcast <- Message{Type: "server_ack", ID: msg.ID, ReceiverUsername: msg.Username}
					}
				} else {
					_, dbErr := database.DB.Exec(
						"INSERT INTO messages (username, receiver_username, content, client_msg_id, status, created_at) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)",
						msg.Username, msg.ReceiverUsername, msg.Content, msg.ID, "sent",
					)
					if dbErr == nil {
						broadcast <- Message{Type: "server_ack", ID: msg.ID, ReceiverUsername: msg.Username}
					}
				}
			} else if msg.Type == "delivery_ack" {
				database.DB.Exec("UPDATE messages SET status = 'delivered' WHERE client_msg_id = $1 AND status = 'sent'", msg.ID)
			} else if msg.Type == "read_ack" {
				_, dbErr := database.DB.Exec(
					"UPDATE messages SET status = 'read' WHERE username = $1 AND receiver_username = $2 AND status IN ('sent', 'delivered')",
					msg.ReceiverUsername, username,
				)
				if dbErr == nil {
					broadcast <- Message{Type: "read_ack", Username: username, ReceiverUsername: msg.ReceiverUsername}
				}
			}
			broadcast <- msg
		}
	}
}

// HandleMessages acts as a global router, distributing broadcasted messages to the correct WebSocket clients.
func HandleMessages() {
	for {
		msg := <-broadcast
		msgBytes, _ := json.Marshal(msg)

		var groupMembers map[string]bool
		if msg.Type == "chat" && msg.GroupID != 0 {
			groupMembers = make(map[string]bool)
			rows, _ := database.DB.Query("SELECT username FROM group_members WHERE group_id = $1", msg.GroupID)
			for rows.Next() {
				var u string
				rows.Scan(&u)
				groupMembers[u] = true
			}
			rows.Close()
		}

		mutex.RLock()
		for client, clientUsername := range clients {
			shouldSend := false

			if msg.Type == "status" || msg.Type == "typing" || msg.Type == "server_ack" || msg.Type == "delivery_ack" || msg.Type == "read_ack" {
				if msg.Type == "status" || clientUsername == msg.ReceiverUsername || clientUsername == msg.Username {
					shouldSend = true
				}
			} else if msg.Type == "chat" {
				if msg.GroupID != 0 {
					if groupMembers[clientUsername] {
						shouldSend = true
					}
				} else {
					if clientUsername == msg.ReceiverUsername || clientUsername == msg.Username {
						shouldSend = true
					}
				}
			}

			if shouldSend {
				client.WriteMessage(websocket.TextMessage, msgBytes)
			}
		}
		mutex.RUnlock()
	}
}