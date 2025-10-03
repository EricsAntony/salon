package model

import "time"

type Salon struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	Contact            map[string]any     `json:"contact,omitempty"`
	Address            map[string]any     `json:"address,omitempty"`
	GeoLocation        map[string]any     `json:"geo_location,omitempty"`
	Logo               *string            `json:"logo,omitempty"`
	Banner             *string            `json:"banner,omitempty"`
	WorkingHours       map[string]any     `json:"working_hours,omitempty"`
	Holidays           map[string]any     `json:"holidays,omitempty"`
	CancellationPolicy *string            `json:"cancellation_policy,omitempty"`
	PaymentModes       []string           `json:"payment_modes"`
	DefaultCurrency    string             `json:"default_currency"`
	TaxRate            float64            `json:"tax_rate"`
	Settings           map[string]any     `json:"settings,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type Branch struct {
	ID           string             `json:"id"`
	SalonID      string             `json:"salon_id"`
	Name         string             `json:"name"`
	Address      map[string]any     `json:"address,omitempty"`
	GeoLocation  map[string]any     `json:"geo_location,omitempty"`
	WorkingHours map[string]any     `json:"working_hours,omitempty"`
	Holidays     map[string]any     `json:"holidays,omitempty"`
	Images       []string           `json:"images"`
	Contact      map[string]any     `json:"contact,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type Category struct {
	ID          string    `json:"id"`
	SalonID     string    `json:"salon_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServiceStatus string

const (
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusInactive ServiceStatus = "inactive"
)

type Service struct {
	ID          string        `json:"id"`
	SalonID     string        `json:"salon_id"`
	CategoryID  string        `json:"category_id"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	DurationMin int           `json:"duration_minutes"`
	Price       float64       `json:"price"`
	Tags        []string      `json:"tags"`
	Status      ServiceStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type StaffStatus string

const (
	StaffStatusActive   StaffStatus = "active"
	StaffStatusInactive StaffStatus = "inactive"
)

type Staff struct {
	ID             string             `json:"id"`
	SalonID        string             `json:"salon_id"`
	Name           string             `json:"name"`
	PhoneNumber    string             `json:"phone_number"`
	Email          *string            `json:"email,omitempty"`
	Role           *string            `json:"role,omitempty"`
	Specialization *string            `json:"specialization,omitempty"`
	Photo          *string            `json:"photo,omitempty"`
	Status         StaffStatus        `json:"status"`
	Shifts         map[string]any     `json:"shifts,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

type StaffService struct {
	ID        string `json:"id"`
	StaffID   string `json:"staff_id"`
	ServiceID string `json:"service_id"`
}

type StaffOTP struct {
	ID          int64     `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	CodeHash    string    `json:"-"`
	ExpiresAt   time.Time `json:"expires_at"`
	Attempts    int       `json:"attempts"`
	CreatedAt   time.Time `json:"created_at"`
}

type StaffRefreshToken struct {
	ID        int64     `json:"id"`
	StaffID   string    `json:"staff_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}
