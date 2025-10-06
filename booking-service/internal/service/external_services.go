package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ExternalService defines interfaces for external service integrations
type ExternalService interface {
	UserService
	SalonService
	PaymentService
	NotificationService
}

// UserService interface for user-service integration
type UserService interface {
	ValidateUser(ctx context.Context, userID uuid.UUID) (*UserInfo, error)
}

// SalonService interface for salon-service integration
type SalonService interface {
	GetSalon(ctx context.Context, salonID uuid.UUID) (*SalonInfo, error)
	GetBranch(ctx context.Context, salonID, branchID uuid.UUID) (*BranchInfo, error)
	GetService(ctx context.Context, salonID, serviceID uuid.UUID) (*ServiceInfo, error)
	GetStylist(ctx context.Context, salonID, stylistID uuid.UUID) (*StylistInfo, error)
	GetStylistSchedule(ctx context.Context, salonID, stylistID uuid.UUID, date time.Time) (*StylistSchedule, error)
	GetStylistServices(ctx context.Context, salonID, stylistID uuid.UUID) ([]*ServiceInfo, error)
}

// PaymentService interface for payment processing (placeholder)
type PaymentService interface {
	ProcessPayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error)
}

// NotificationService interface for notifications (placeholder)
type NotificationService interface {
	SendBookingConfirmation(ctx context.Context, booking *BookingNotification) error
	SendBookingCancellation(ctx context.Context, booking *BookingNotification) error
	SendBookingReschedule(ctx context.Context, booking *BookingNotification) error
	SendBookingReminder(ctx context.Context, booking *BookingNotification) error
}

// External service data structures
type UserInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Phone string    `json:"phone"`
}

type SalonInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
}

type BranchInfo struct {
	ID       uuid.UUID `json:"id"`
	SalonID  uuid.UUID `json:"salon_id"`
	Name     string    `json:"name"`
	Address  string    `json:"address"`
	Phone    string    `json:"phone"`
	Timezone string    `json:"timezone"`
}

type ServiceInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"` // in minutes
	Price       float64   `json:"price"`
	CategoryID  uuid.UUID `json:"category_id"`
}

type StylistInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	BranchID uuid.UUID `json:"branch_id"`
	Phone    string    `json:"phone"`
	Email    string    `json:"email"`
}

type StylistSchedule struct {
	StylistID    uuid.UUID     `json:"stylist_id"`
	Date         time.Time     `json:"date"`
	WorkingHours []WorkingHour `json:"working_hours"`
	Breaks       []BreakPeriod `json:"breaks"`
}

type WorkingHour struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type BreakPeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Type      string    `json:"type"` // "lunch", "break", etc.
}

// Payment service structures (placeholder)
type PaymentRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	UserID    uuid.UUID `json:"user_id"`
	Method    string    `json:"method"`
}

type PaymentResponse struct {
	PaymentID     string    `json:"payment_id"`
	Status        string    `json:"status"`
	TransactionID string    `json:"transaction_id"`
	ProcessedAt   time.Time `json:"processed_at"`
}

type RefundResponse struct {
	RefundID    string    `json:"refund_id"`
	Status      string    `json:"status"`
	Amount      float64   `json:"amount"`
	ProcessedAt time.Time `json:"processed_at"`
}

// Notification service structures (placeholder)
type BookingNotification struct {
	BookingID   uuid.UUID `json:"booking_id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	UserPhone   string    `json:"user_phone"`
	SalonName   string    `json:"salon_name"`
	BranchName  string    `json:"branch_name"`
	Services    []string  `json:"services"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TotalAmount float64   `json:"total_amount"`
}

// Implementation of external services
type externalService struct {
	userServiceURL  string
	salonServiceURL string
	httpClient      *http.Client
}

// NewExternalService creates a new external service client
func NewExternalService(userServiceURL, salonServiceURL string) ExternalService {
	return &externalService{
		userServiceURL:  userServiceURL,
		salonServiceURL: salonServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateUser validates a user exists and returns user info
func (e *externalService) ValidateUser(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	url := fmt.Sprintf("%s/user/%s", e.userServiceURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

// GetSalon retrieves salon information
func (e *externalService) GetSalon(ctx context.Context, salonID uuid.UUID) (*SalonInfo, error) {
	url := fmt.Sprintf("%s/salons/%s", e.salonServiceURL, salonID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call salon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("salon not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("salon service returned status %d", resp.StatusCode)
	}

	var salon SalonInfo
	if err := json.NewDecoder(resp.Body).Decode(&salon); err != nil {
		return nil, fmt.Errorf("failed to decode salon response: %w", err)
	}

	return &salon, nil
}

// GetBranch retrieves branch information
func (e *externalService) GetBranch(ctx context.Context, salonID, branchID uuid.UUID) (*BranchInfo, error) {
	url := fmt.Sprintf("%s/salons/%s/branches/%s", e.salonServiceURL, salonID, branchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call salon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("branch not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("salon service returned status %d", resp.StatusCode)
	}

	var branch BranchInfo
	if err := json.NewDecoder(resp.Body).Decode(&branch); err != nil {
		return nil, fmt.Errorf("failed to decode branch response: %w", err)
	}

	return &branch, nil
}

// GetService retrieves service information
func (e *externalService) GetService(ctx context.Context, salonID, serviceID uuid.UUID) (*ServiceInfo, error) {
	url := fmt.Sprintf("%s/salons/%s/services/%s", e.salonServiceURL, salonID, serviceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call salon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("service not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("salon service returned status %d", resp.StatusCode)
	}

	var service ServiceInfo
	if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
		return nil, fmt.Errorf("failed to decode service response: %w", err)
	}

	return &service, nil
}

// GetStylist retrieves stylist information
func (e *externalService) GetStylist(ctx context.Context, salonID, stylistID uuid.UUID) (*StylistInfo, error) {
	url := fmt.Sprintf("%s/salons/%s/staff/%s", e.salonServiceURL, salonID, stylistID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call salon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("stylist not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("salon service returned status %d", resp.StatusCode)
	}

	var stylist StylistInfo
	if err := json.NewDecoder(resp.Body).Decode(&stylist); err != nil {
		return nil, fmt.Errorf("failed to decode stylist response: %w", err)
	}

	return &stylist, nil
}

// GetStylistSchedule retrieves stylist schedule for a specific date
func (e *externalService) GetStylistSchedule(ctx context.Context, salonID, stylistID uuid.UUID, date time.Time) (*StylistSchedule, error) {
	url := fmt.Sprintf("%s/salons/%s/staff/%s/schedule?date=%s", e.salonServiceURL, salonID, stylistID, date.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call salon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("stylist schedule not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("salon service returned status %d", resp.StatusCode)
	}

	var schedule StylistSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode schedule response: %w", err)
	}

	return &schedule, nil
}

// GetStylistServices retrieves services offered by a stylist
func (e *externalService) GetStylistServices(ctx context.Context, salonID, stylistID uuid.UUID) ([]*ServiceInfo, error) {
	url := fmt.Sprintf("%s/salons/%s/staff/%s/services", e.salonServiceURL, salonID, stylistID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call salon service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("salon service returned status %d", resp.StatusCode)
	}

	var services []*ServiceInfo
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, fmt.Errorf("failed to decode services response: %w", err)
	}

	return services, nil
}

// Placeholder implementations for payment and notification services
func (e *externalService) ProcessPayment(ctx context.Context, request *PaymentRequest) (*PaymentResponse, error) {
	// TODO: Implement actual payment processing
	return &PaymentResponse{
		PaymentID:     fmt.Sprintf("pay_%s", uuid.New().String()[:8]),
		Status:        "success",
		TransactionID: fmt.Sprintf("txn_%s", uuid.New().String()[:8]),
		ProcessedAt:   time.Now(),
	}, nil
}

func (e *externalService) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	// TODO: Implement actual refund processing
	return &RefundResponse{
		RefundID:    fmt.Sprintf("ref_%s", uuid.New().String()[:8]),
		Status:      "success",
		Amount:      amount,
		ProcessedAt: time.Now(),
	}, nil
}

func (e *externalService) SendBookingConfirmation(ctx context.Context, booking *BookingNotification) error {
	// TODO: Implement actual notification sending
	return nil
}

func (e *externalService) SendBookingCancellation(ctx context.Context, booking *BookingNotification) error {
	// TODO: Implement actual notification sending
	return nil
}

func (e *externalService) SendBookingReschedule(ctx context.Context, booking *BookingNotification) error {
	// TODO: Implement actual notification sending
	return nil
}

func (e *externalService) SendBookingReminder(ctx context.Context, booking *BookingNotification) error {
	// TODO: Implement actual notification sending
	return nil
}
