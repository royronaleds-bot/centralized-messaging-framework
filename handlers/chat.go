package handlers

import (
	"corehub/database"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]string)
	broadcast = make(chan Message)
	mutex     = &sync.RWMutex{}
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type Message struct {
	Username         string `json:"username"`          // اسم المرسل
	ReceiverUsername string `json:"receiver_username"` // اسم المستلم (الجديد)
	Content          string `json:"content"`           // محتوى الرسالة
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// جلب اسم المستخدم الذي وضعه الميدل وير في الـ Context
	val := r.Context().Value("username")
	if val == nil {
		log.Println("⚠️ فشل الاتصال: اسم المستخدم غير موجود في السياق")
		return
	}
	username := val.(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ خطأ في ترقية الاتصال: %v", err)
		return
	}

	mutex.Lock()
	clients[conn] = username
	mutex.Unlock()

	log.Printf("✅ متصل الآن: %s", username)

	defer func() {
		mutex.Lock()
		delete(clients, conn)
		mutex.Unlock()
		conn.Close()
		log.Printf("🔌 انقطع اتصال: %s", username)
	}()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(rawMsg, &msg); err == nil {
			msg.Username = username

			// حفظ الرسالة في قاعدة البيانات
			_, dbErr := database.DB.Exec("INSERT INTO messages (username, receiver_username, content) VALUES ($1, $2, $3)", msg.Username, msg.ReceiverUsername, msg.Content)
			if dbErr != nil {
				log.Printf("❌ خطأ في حفظ الرسالة: %v", dbErr)
			}

			log.Printf("📩 رسالة من [%s]: %s", msg.Username, msg.Content)
			broadcast <- msg
		}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		msgBytes, _ := json.Marshal(msg)

		mutex.RLock()
		for client, clientUsername := range clients {
			// نرسل الرسالة فقط إذا كان المتصل هو "المستلم" أو "المرسل نفسه"
			// (نرسلها للمرسل أيضاً لكي تظهر لديه على الشاشة كدليل على نجاح الإرسال)
			if clientUsername == msg.ReceiverUsername || clientUsername == msg.Username {
				err := client.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					log.Printf("❌ فشل الإرسال لعميل: %v", err)
					client.Close()
					// هنا نحتاج إلى حذف العميل من قائمة clients،
					// لكن لا يمكننا فعل ذلك داخل RLock، سيتم حذفه تلقائياً من HandleConnections
				}
			}
		}
		mutex.RUnlock()
	}
}

// دالة لجلب الرسائل القديمة من قاعدة البيانات
// دالة لجلب الرسائل القديمة بين شخصين محددين
func GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 1. جلب اسمي أنا (من الميدل وير)
	val := r.Context().Value("username")
	if val == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	myUsername := val.(string)

	// 2. جلب اسم الشخص الذي أريد التحدث معه (من الرابط)
	otherUser := r.URL.Query().Get("with")
	if otherUser == "" {
		json.NewEncoder(w).Encode([]Message{})
		return
	}

	// 3. استعلام ذكي يجلب الرسائل المتبادلة بيننا فقط
	query := `
		SELECT username, receiver_username, content 
		FROM messages 
		WHERE (username = $1 AND receiver_username = $2) 
		   OR (username = $2 AND receiver_username = $1) 
		ORDER BY id ASC
	`
	rows, err := database.DB.Query(query, myUsername, otherUser)
	if err != nil {
		log.Printf("❌ خطأ في جلب الرسائل: %v", err)
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Username, &msg.ReceiverUsername, &msg.Content); err == nil {
			msgs = append(msgs, msg)
		}
	}

	json.NewEncoder(w).Encode(msgs)
}
