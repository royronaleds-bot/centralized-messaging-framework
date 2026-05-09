package main

import (
	"corehub/database"
	"corehub/handlers"
	"corehub/middleware"
	"log"
	"net/http"
)

func main() {
	database.Connect()

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/ws", middleware.AuthMiddleware(handlers.HandleConnections))

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
