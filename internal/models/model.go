package models

import (
	"time"
)

type Token struct {
	Token string `json:"token"`
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type Reception struct {
	ID       string     `json:"id"`
	DateTime time.Time  `json:"dateTime"`
	PvzId    string     `json:"pvzId"`
	Status   string     `json:"status"`
	ClosedAt *time.Time `json:"closedAt"`
}

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionId string    `json:"receptionId"`
}

type Error struct {
	Message string `json:"message"`
}
