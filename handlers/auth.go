package handlers

import (
	"corehub/database"
	"corehub/models"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	_, err := database.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User %s registered successfully", user.Username)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	var storedPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE username=$1", user.Username).Scan(&storedPassword)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)) != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "Login successful! Welcome %s", user.Username)
}
