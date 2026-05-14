package main

import (
	"corehub/database"
	"corehub/handlers"
	"corehub/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// تم تعديل الاسم ليطابق الدالة الموجودة في db.go
	database.Connect()

	r := mux.NewRouter()
	// مسار جلب الرسائل القديمة (محمي بالمصادقة)
	r.HandleFunc("/messages", middleware.AuthMiddleware(handlers.GetMessages)).Methods("GET")

	// تم تعديل الأسماء لتطابق الدوال الموجودة في auth.go
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// مسار الويب سوكيت محمي بالميدل وير
	r.HandleFunc("/ws", middleware.AuthMiddleware(handlers.HandleConnections))

	// تقديم ملفات الـ HTML
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))

	// تشغيل البث في الخلفية
	go handlers.HandleMessages()

	log.Println("🚀 السيرفر انطلق على: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
