package handlers

import (
	"corehub/database"
	"encoding/json"
	"log"
	"net/http"
	"sync"

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

type Message struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Username         string `json:"username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
	Status           string `json:"status"` // أضفنا هذا الحقل لمزامنة الحالة مع القاعدة
}

type UserWithStatus struct {
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

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
	mutex.Unlock()

	broadcast <- Message{Type: "status", Username: username, Content: "online"}

	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
		broadcast <- Message{Type: "status", Username: username, Content: "offline"}
	}()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(rawMsg, &msg); err == nil {
			msg.Username = username

			if msg.Type == "chat" {
				// حفظ الرسالة مع الـ ID القادم من الواجهة والحالة الافتراضية 'sent'
				_, dbErr := database.DB.Exec(
					"INSERT INTO messages (username, receiver_username, content, client_msg_id, status) VALUES ($1, $2, $3, $4, $5)",
					msg.Username, msg.ReceiverUsername, msg.Content, msg.ID, "sent",
				)
				if dbErr == nil {
					broadcast <- Message{Type: "server_ack", ID: msg.ID, ReceiverUsername: msg.Username}
				}
			} else if msg.Type == "delivery_ack" {
				// --- التحديث الجديد: حفظ حالة التسليم في قاعدة البيانات ---
				_, dbErr := database.DB.Exec(
					"UPDATE messages SET status = 'delivered' WHERE client_msg_id = $1",
					msg.ID,
				)
				if dbErr != nil {
					log.Printf("❌ فشل تحديث حالة التسليم: %v", dbErr)
				}
			}

			broadcast <- msg
		}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		msgBytes, _ := json.Marshal(msg)
		mutex.RLock()
		for client, clientUsername := range clients {
			if msg.Type == "status" || msg.Type == "typing" || msg.Type == "server_ack" || msg.Type == "delivery_ack" || clientUsername == msg.ReceiverUsername || clientUsername == msg.Username {
				client.WriteMessage(websocket.TextMessage, msgBytes)
			}
		}
		mutex.RUnlock()
	}
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)
	otherUser := r.URL.Query().Get("with")

	// --- التحديث الجديد: جلب الـ ID والحالة الحقيقية من القاعدة ---
	query := `
		SELECT username, receiver_username, content, client_msg_id, status 
		FROM messages 
		WHERE (username = $1 AND receiver_username = $2) 
		   OR (username = $2 AND receiver_username = $1) 
		ORDER BY id ASC
	`
	rows, err := database.DB.Query(query, myUsername, otherUser)
	if err != nil {
		http.Error(w, "Error", 500)
		return
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Username, &m.ReceiverUsername, &m.Content, &m.ID, &m.Status); err == nil {
			msgs = append(msgs, m)
		}
	}
	json.NewEncoder(w).Encode(msgs)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)
	mutex.RLock()
	onlineMap := make(map[string]bool)
	for _, u := range clients {
		onlineMap[u] = true
	}
	mutex.RUnlock()

	rows, _ := database.DB.Query("SELECT username FROM users WHERE username != $1", myUsername)
	defer rows.Close()
	var users []UserWithStatus
	for rows.Next() {
		var u string
		rows.Scan(&u)
		users = append(users, UserWithStatus{Username: u, Online: onlineMap[u]})
	}
	json.NewEncoder(w).Encode(users)
}
