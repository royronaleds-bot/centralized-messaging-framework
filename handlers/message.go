package handlers

import (
	"corehub/database"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

// GetMessages fetches the message history for a specific 1-on-1 chat or group chat.
func GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)

	otherUser := r.URL.Query().Get("with")
	groupID := r.URL.Query().Get("group_id")

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	var rows *sql.Rows
	var err error

	if groupID != "" {
		query := `
			SELECT username, COALESCE(receiver_username, ''), content, COALESCE(client_msg_id, ''), COALESCE(status, 'sent'), COALESCE(group_id, 0), COALESCE(TO_CHAR(created_at, 'HH12:MI AM'), '')
			FROM messages 
			WHERE group_id = $1 
			ORDER BY id DESC LIMIT 50 OFFSET $2
		`
		rows, err = database.DB.Query(query, groupID, offset)
	} else if otherUser != "" {
		query := `
			SELECT username, COALESCE(receiver_username, ''), content, COALESCE(client_msg_id, ''), COALESCE(status, 'sent'), COALESCE(group_id, 0), COALESCE(TO_CHAR(created_at, 'HH12:MI AM'), '')
			FROM messages 
			WHERE (username = $1 AND receiver_username = $2) 
			   OR (username = $2 AND receiver_username = $1) 
			ORDER BY id DESC LIMIT 50 OFFSET $3
		`
		rows, err = database.DB.Query(query, myUsername, otherUser, offset)
	} else {
		json.NewEncoder(w).Encode([]Message{})
		return
	}

	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Username, &m.ReceiverUsername, &m.Content, &m.ID, &m.Status, &m.GroupID, &m.CreatedAt); err == nil {
			msgs = append(msgs, m)
		}
	}

	// Reverse slice to show oldest to newest in the UI
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	json.NewEncoder(w).Encode(msgs)
}
