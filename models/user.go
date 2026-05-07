package models

// User struct for authentication data
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
