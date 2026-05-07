package main

import (
	"corehub/database"
	"corehub/handlers"
	"fmt"
	"net/http"
)

func main() {
	// 1. Initialize Database
	database.InitDB()

	// 2. Define Routes
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/ws", handlers.HandleConnections)

	// 3. Start Server
	fmt.Println("CoreHub Server is live on :8080")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
