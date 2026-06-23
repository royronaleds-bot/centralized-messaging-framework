package main

import (
	"corehub/database"
	"corehub/handlers"
	"corehub/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// main is the entry point for the CoreHub server.
// It establishes the database connection, configures routing, and starts the HTTP server.
func main() {
	// 1. Initialize Database Connection
	database.Connect()

	// 2. Initialize Router
	r := mux.NewRouter()

	// =========================================================
	// 📌 PUBLIC API ROUTES
	// =========================================================
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// =========================================================
	// 🔒 PROTECTED API ROUTES (Requires JWT Token)
	// =========================================================

	// --- Chat & Users ---
	r.HandleFunc("/messages", middleware.AuthMiddleware(handlers.GetMessages)).Methods("GET")
	r.HandleFunc("/users", middleware.AuthMiddleware(handlers.GetUsers)).Methods("GET")
	r.HandleFunc("/search", middleware.AuthMiddleware(handlers.SearchUsers)).Methods("GET")

	// --- Groups ---
	r.HandleFunc("/groups", middleware.AuthMiddleware(handlers.GetGroups)).Methods("GET")
	r.HandleFunc("/groups/join", middleware.AuthMiddleware(handlers.JoinGroup)).Methods("POST")

	// --- Admin Controls ---
	r.HandleFunc("/ban", middleware.AuthMiddleware(handlers.BanUser)).Methods("POST")
	r.HandleFunc("/unban", middleware.AuthMiddleware(handlers.UnbanUser)).Methods("POST")
	r.HandleFunc("/stats", middleware.AuthMiddleware(handlers.GetStats)).Methods("GET")

	// =========================================================
	// 📁 FILE UPLOADS & STATIC ASSETS
	// =========================================================
	r.HandleFunc("/upload", middleware.AuthMiddleware(handlers.UploadHandler)).Methods("POST")
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// =========================================================
	// 🔌 WEBSOCKETS & FRONTEND SERVING
	// =========================================================
	r.HandleFunc("/ws", middleware.AuthMiddleware(handlers.HandleConnections))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))

	// 3. Start Background Message Dispatcher (Goroutine)
	go handlers.HandleMessages()

	// 4. Start Server
	log.Println("🚀 Server successfully launched on Port 80 (Ready for LAN connections)")
	log.Fatal(http.ListenAndServe(":80", r))
}
