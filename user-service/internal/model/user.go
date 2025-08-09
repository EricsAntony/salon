package model

import (
	"time"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type User struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Name        string    `json:"name"`
	Gender      Gender    `json:"gender"`
	Email       *string   `json:"email,omitempty"`
	Location    *string   `json:"location,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        int64
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

type OTP struct {
	ID          int64
	PhoneNumber string
	CodeHash    string
	ExpiresAt   time.Time
	Attempts    int
	CreatedAt   time.Time
}
