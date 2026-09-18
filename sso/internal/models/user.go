package models

type User struct {
	ID           int64
	Email        string
	IsAdmin      bool 
	PasswordHash string
}
