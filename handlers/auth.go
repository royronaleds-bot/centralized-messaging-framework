package handlers

import (
	"corehub/database"
	"corehub/utils"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// User represents the payload for registration and login requests.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Secret   string `json:"secret"`
}

// RegisterHandler handles new user registration and assigns admin roles based on a secret key.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var user User
	json.NewDecoder(r.Body).Decode(&user)

	role := "user"
	if user.Secret == "EngCore2026" {
		role = "admin"
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	_, err := database.DB.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", user.Username, hashedPassword, role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Registration failed"})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully", "role": role})
}

// LoginHandler verifies user credentials, checks for active bans, and returns a JWT token.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var user User
	json.NewDecoder(r.Body).Decode(&user)

	var hashedPassword string
	var isBanned bool

	err := database.DB.QueryRow("SELECT password, COALESCE(is_banned, false) FROM users WHERE username=$1", user.Username).Scan(&hashedPassword, &isBanned)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
		return
	}

	// 🚫 منع الدخول إذا كان الحساب محظوراً
	if isBanned {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "Your account has been banned by an Administrator."})
		return
	}

	token, _ := utils.GenerateToken(user.Username)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
