package handlers

import (
	"corehub/database"
	"database/sql"
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

type Message struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Username         string `json:"username"`
	ReceiverUsername string `json:"receiver_username"`
	GroupID          int    `json:"group_id"`
	Content          string `json:"content"`
	Status           string `json:"status"`
}

type UserWithStatus struct {
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
				if msg.GroupID != 0 {
					_, dbErr := database.DB.Exec(
						"INSERT INTO messages (username, content, client_msg_id, status, group_id) VALUES ($1, $2, $3, $4, $5)",
						msg.Username, msg.Content, msg.ID, "sent", msg.GroupID,
					)
					if dbErr == nil {
						broadcast <- Message{Type: "server_ack", ID: msg.ID, ReceiverUsername: msg.Username}
					} else {
						log.Println("❌ DB Insert Group Msg Error:", dbErr)
					}
				} else {
					_, dbErr := database.DB.Exec(
						"INSERT INTO messages (username, receiver_username, content, client_msg_id, status) VALUES ($1, $2, $3, $4, $5)",
						msg.Username, msg.ReceiverUsername, msg.Content, msg.ID, "sent",
					)
					if dbErr == nil {
						broadcast <- Message{Type: "server_ack", ID: msg.ID, ReceiverUsername: msg.Username}
					} else {
						log.Println("❌ DB Insert Private Msg Error:", dbErr)
					}
				}
			} else if msg.Type == "delivery_ack" {
				_, dbErr := database.DB.Exec("UPDATE messages SET status = 'delivered' WHERE client_msg_id = $1", msg.ID)
				if dbErr != nil {
					log.Println("❌ DB Update Delivery Error:", dbErr)
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

			if msg.Type == "status" || msg.Type == "typing" || msg.Type == "server_ack" || msg.Type == "delivery_ack" {
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

func GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)

	otherUser := r.URL.Query().Get("with")
	groupID := r.URL.Query().Get("group_id")

	var rows *sql.Rows
	var err error

	if groupID != "" {
		query := `
			SELECT username, COALESCE(receiver_username, ''), content, COALESCE(client_msg_id, ''), COALESCE(status, 'sent'), COALESCE(group_id, 0)
			FROM messages 
			WHERE group_id = $1 
			ORDER BY id ASC
		`
		rows, err = database.DB.Query(query, groupID)
	} else if otherUser != "" {
		query := `
			SELECT username, COALESCE(receiver_username, ''), content, COALESCE(client_msg_id, ''), COALESCE(status, 'sent'), COALESCE(group_id, 0)
			FROM messages 
			WHERE (username = $1 AND receiver_username = $2) 
			   OR (username = $2 AND receiver_username = $1) 
			ORDER BY id ASC
		`
		rows, err = database.DB.Query(query, myUsername, otherUser)
	} else {
		json.NewEncoder(w).Encode([]Message{})
		return
	}

	if err != nil {
		// 🔴 هنا سيكتب السيرفر سبب الانهيار في شاشة الـ Terminal
		log.Println("❌ DB Query Error in GetMessages:", err)
		http.Error(w, "Error fetching messages", 500)
		return
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Username, &m.ReceiverUsername, &m.Content, &m.ID, &m.Status, &m.GroupID); err != nil {
			// 🔴 هنا سيكشف السيرفر إذا كان هناك مشكلة في قراءة أحد الأعمدة
			log.Println("❌ Row Scan Error in GetMessages:", err)
		} else {
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

func JoinGroup(w http.ResponseWriter, r *http.Request) {
	myUsername := r.Context().Value("username").(string)
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Name == "" {
		http.Error(w, "Group name required", 400)
		return
	}

	var groupID int
	err := database.DB.QueryRow("SELECT id FROM groups WHERE name = $1", req.Name).Scan(&groupID)
	if err != nil {
		err = database.DB.QueryRow("INSERT INTO groups (name) VALUES ($1) RETURNING id", req.Name).Scan(&groupID)
		if err != nil {
			http.Error(w, "Failed to create group", 500)
			return
		}
	}

	database.DB.Exec("INSERT INTO group_members (group_id, username) VALUES ($1, $2) ON CONFLICT DO NOTHING", groupID, myUsername)
	json.NewEncoder(w).Encode(Group{ID: groupID, Name: req.Name})
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)

	rows, err := database.DB.Query(`
		SELECT g.id, g.name 
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.username = $1
	`, myUsername)

	if err != nil {
		http.Error(w, "Failed to fetch groups", 500)
		return
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		rows.Scan(&g.ID, &g.Name)
		groups = append(groups, g)
	}
	json.NewEncoder(w).Encode(groups)
}
