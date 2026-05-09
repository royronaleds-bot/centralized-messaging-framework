package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// DB must start with Uppercase
var DB *sql.DB

func Connect() {
	var err error
	// تأكد من password و dbname هنا
	connStr := "host=localhost port=5432 user=postgres password=123456 dbname=corehub_db sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}
	log.Println("Successfully connected to database!")
}
