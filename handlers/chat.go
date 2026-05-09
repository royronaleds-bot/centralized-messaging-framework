package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]string)
var mu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = username
	mu.Unlock()

	broadcast(fmt.Sprintf("System: %s joined", username))

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			broadcast(fmt.Sprintf("System: %s left", username))
			break
		}
		broadcast(fmt.Sprintf("%s: %s", username, string(msg)))
	}
}

func broadcast(message string) {
	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		client.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
