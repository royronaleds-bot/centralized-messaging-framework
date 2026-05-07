package handlers

import (
	"corehub/database"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	fmt.Printf("New Client Connected! Total Active: %d\n", len(clients))
	clientsMu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			fmt.Printf("Client Disconnected. Remaining: %d\n", len(clients))
			clientsMu.Unlock()
			break
		}

		database.DB.Exec("INSERT INTO messages (sender, content) VALUES ($1, $2)", "User", string(msg))
		broadcastMessage(string(msg))
	}
}

func broadcastMessage(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
