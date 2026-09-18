package models

type App struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}
