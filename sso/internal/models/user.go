package models

type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	IsAdmin      bool   `json:"is_admin"`
	PasswordHash string `json:"password_hash"`
}
