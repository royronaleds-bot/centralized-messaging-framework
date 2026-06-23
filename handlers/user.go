package handlers

import (
	"corehub/database"
	"encoding/json"
	"net/http"
	"runtime"
	"sort"

	"github.com/gorilla/websocket"
)

// UserWithStatus represents a user alongside their current connection and ban status.
type UserWithStatus struct {
	Username    string `json:"username"`
	Online      bool   `json:"online"`
	LastMessage string `json:"last_message"`
	LastTime    string `json:"last_time"`
	Role        string `json:"role"`
	IsBanned    bool   `json:"is_banned"`
}

// BanUser restricts a user from logging in and terminates their active WebSocket connection.
func BanUser(w http.ResponseWriter, r *http.Request) {
	myUsername := r.Context().Value("username").(string)
	var role string
	database.DB.QueryRow("SELECT role FROM users WHERE username = $1", myUsername).Scan(&role)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	database.DB.Exec("UPDATE users SET is_banned = true WHERE username = $1", req.Username)

	mutex.Lock()
	for conn, uname := range clients {
		if uname == req.Username {
			banMsg := Message{Type: "banned"}
			msgBytes, _ := json.Marshal(banMsg)
			conn.WriteMessage(websocket.TextMessage, msgBytes)
			conn.Close()
			delete(clients, conn)
		}
	}
	mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"status": "banned"})
}

// UnbanUser restores access to a previously banned user.
func UnbanUser(w http.ResponseWriter, r *http.Request) {
	myUsername := r.Context().Value("username").(string)
	var role string
	database.DB.QueryRow("SELECT role FROM users WHERE username = $1", myUsername).Scan(&role)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	database.DB.Exec("UPDATE users SET is_banned = false WHERE username = $1", req.Username)
	json.NewEncoder(w).Encode(map[string]string{"status": "unbanned"})
}

// GetStats returns server metrics (CPU cores, Memory usage, active Goroutines).
func GetStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats := map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  m.Alloc / 1024 / 1024,
		"cpu_cores":  runtime.NumCPU(),
	}
	json.NewEncoder(w).Encode(stats)
}

// GetUsers retrieves the list of users based on the requester's role (Admin vs User).
func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)

	var role string
	err := database.DB.QueryRow("SELECT role FROM users WHERE username = $1", myUsername).Scan(&role)
	if err != nil {
		role = "user"
	}

	mutex.RLock()
	onlineMap := make(map[string]bool)
	for _, u := range clients {
		onlineMap[u] = true
	}
	mutex.RUnlock()

	baseQuery := `
		SELECT u.username, u.role, COALESCE(u.is_banned, false),
			COALESCE((SELECT content FROM messages 
					  WHERE (username = $1 AND receiver_username = u.username) 
					     OR (username = u.username AND receiver_username = $1) 
					  ORDER BY created_at DESC LIMIT 1), ''),
			COALESCE((SELECT TO_CHAR(created_at, 'HH12:MI AM') FROM messages 
					  WHERE (username = $1 AND receiver_username = u.username) 
					     OR (username = u.username AND receiver_username = $1) 
					  ORDER BY created_at DESC LIMIT 1), '')
		FROM users u
		WHERE u.username != $1
	`

	if role == "user" {
		baseQuery += ` AND COALESCE(u.is_banned, false) = false AND EXISTS (SELECT 1 FROM messages m WHERE (m.username = u.username AND m.receiver_username = $1) OR (m.username = $1 AND m.receiver_username = u.username))`
	}

	baseQuery += ` ORDER BY COALESCE((SELECT created_at FROM messages WHERE (username = u.username AND receiver_username = $1) OR (username = $1 AND receiver_username = u.username) ORDER BY created_at DESC LIMIT 1), '1970-01-01'::timestamp) DESC`

	rows, dbErr := database.DB.Query(baseQuery, myUsername)
	if dbErr != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserWithStatus
	for rows.Next() {
		var u UserWithStatus
		rows.Scan(&u.Username, &u.Role, &u.IsBanned, &u.LastMessage, &u.LastTime)
		u.Online = onlineMap[u.Username]
		users = append(users, u)
	}

	if role == "admin" {
		sort.SliceStable(users, func(i, j int) bool {
			return users[i].Online && !users[j].Online
		})
	}

	w.Header().Set("X-User-Role", role)
	json.NewEncoder(w).Encode(users)
}

// SearchUsers allows regular users to discover active users globally across the server.
func SearchUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)
	query := r.URL.Query().Get("q")

	if query == "" {
		json.NewEncoder(w).Encode([]UserWithStatus{})
		return
	}

	searchQuery := `SELECT username, role FROM users WHERE username ILIKE $1 AND username != $2 AND COALESCE(is_banned, false) = false LIMIT 10`
	rows, err := database.DB.Query(searchQuery, "%"+query+"%", myUsername)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserWithStatus
	for rows.Next() {
		var u UserWithStatus
		rows.Scan(&u.Username, &u.Role)
		u.LastMessage = "Start a new conversation"
		users = append(users, u)
	}
	json.NewEncoder(w).Encode(users)
}
