package handlers

import (
	"corehub/database"
	"encoding/json"
	"net/http"
)

// Group represents a chat group entity.
// 🌟 هاد هو الهيكل اللي نسيناه وكان مسبب المشكلة
type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	LastMessage string `json:"last_message"`
	LastTime    string `json:"last_time"`
}

// JoinGroup adds a user to a group or creates the group if it does not exist.
func JoinGroup(w http.ResponseWriter, r *http.Request) {
	myUsername := r.Context().Value("username").(string)
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Name == "" {
		http.Error(w, "Group name required", http.StatusBadRequest)
		return
	}

	var groupID int
	err := database.DB.QueryRow("SELECT id FROM groups WHERE name = $1", req.Name).Scan(&groupID)
	if err != nil {
		err = database.DB.QueryRow("INSERT INTO groups (name) VALUES ($1) RETURNING id", req.Name).Scan(&groupID)
		if err != nil {
			http.Error(w, "Failed to create group", http.StatusInternalServerError)
			return
		}
	}

	database.DB.Exec("INSERT INTO group_members (group_id, username) VALUES ($1, $2) ON CONFLICT DO NOTHING", groupID, myUsername)
	json.NewEncoder(w).Encode(Group{ID: groupID, Name: req.Name})
}

// GetGroups retrieves all groups that the current user is a member of.
func GetGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	myUsername := r.Context().Value("username").(string)

	query := `
		SELECT g.id, g.name,
			COALESCE((SELECT content FROM messages WHERE group_id = g.id ORDER BY created_at DESC LIMIT 1), ''),
			COALESCE((SELECT TO_CHAR(created_at, 'HH12:MI AM') FROM messages WHERE group_id = g.id ORDER BY created_at DESC LIMIT 1), '')
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.username = $1
	`
	rows, err := database.DB.Query(query, myUsername)
	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		rows.Scan(&g.ID, &g.Name, &g.LastMessage, &g.LastTime)
		groups = append(groups, g)
	}
	json.NewEncoder(w).Encode(groups)
}
