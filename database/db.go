package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	// Using your credentials: password 123456 and db corehub_db
	connStr := "user=postgres password=123456 dbname=corehub_db sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Database connection error:", err)
		return
	}
	fmt.Println("Database Connected Successfully!")
}
