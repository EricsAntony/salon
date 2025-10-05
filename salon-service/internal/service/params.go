package service

import (
	"strings"

	"salon-service/internal/model"

	sharederrors "github.com/EricsAntony/salon/salon-shared/errors"
	"github.com/google/uuid"
)

type CreateSalonParams struct {
	Name               string
	Description        *string
	Contact            map[string]any
	Address            map[string]any
	GeoLocation        map[string]any
	Logo               *string
	Banner             *string
	WorkingHours       map[string]any
	Holidays           map[string]any
	CancellationPolicy *string
	PaymentModes       []string
	DefaultCurrency    string
	TaxRate            float64
	Settings           map[string]any
}

func (p CreateSalonParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if strings.TrimSpace(p.DefaultCurrency) == "" {
		errs = sharederrors.AppendValidationError(errs, "default_currency", "is required")
	}
	if p.TaxRate < 0 {
		errs = sharederrors.AppendValidationError(errs, "tax_rate", "must be greater than or equal to 0")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateSalonParams struct {
	ID                 string
	Name               string
	Description        *string
	Contact            map[string]any
	Address            map[string]any
	GeoLocation        map[string]any
	Logo               *string
	Banner             *string
	WorkingHours       map[string]any
	Holidays           map[string]any
	CancellationPolicy *string
	PaymentModes       []string
	DefaultCurrency    string
	TaxRate            float64
	Settings           map[string]any
}

func (p UpdateSalonParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.ID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if strings.TrimSpace(p.DefaultCurrency) == "" {
		errs = sharederrors.AppendValidationError(errs, "default_currency", "is required")
	}
	if p.TaxRate < 0 {
		errs = sharederrors.AppendValidationError(errs, "tax_rate", "must be greater than or equal to 0")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type CreateBranchParams struct {
	SalonID      string
	Name         string
	Address      map[string]any
	GeoLocation  map[string]any
	WorkingHours map[string]any
	Holidays     map[string]any
	Images       []string
	Contact      map[string]any
}

func (p CreateBranchParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateBranchParams struct {
	ID           string
	SalonID      string
	Name         string
	Address      map[string]any
	GeoLocation  map[string]any
	WorkingHours map[string]any
	Holidays     map[string]any
	Images       []string
	Contact      map[string]any
}

func (p UpdateBranchParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.ID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "id", "must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type CreateCategoryParams struct {
	SalonID     string
	Name        string
	Description *string
}

func (p CreateCategoryParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateCategoryParams struct {
	ID          string
	SalonID     string
	Name        string
	Description *string
}

func (p UpdateCategoryParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.ID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "id", "must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type CreateServiceParams struct {
	SalonID         string
	CategoryID      string
	Name            string
	Description     *string
	DurationMinutes int
	Price           float64
	Tags            []string
	Status          model.ServiceStatus
}

func (p CreateServiceParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(p.CategoryID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "category_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if p.DurationMinutes <= 0 {
		errs = sharederrors.AppendValidationError(errs, "duration_minutes", "must be greater than 0")
	}
	if p.Price < 0 {
		errs = sharederrors.AppendValidationError(errs, "price", "must be greater than or equal to 0")
	}
	if !isValidServiceStatus(p.Status) {
		errs = sharederrors.AppendValidationError(errs, "status", "must be 'active' or 'inactive'")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateServiceParams struct {
	ID              string
	SalonID         string
	CategoryID      string
	Name            string
	Description     *string
	DurationMinutes int
	Price           float64
	Tags            []string
	Status          model.ServiceStatus
}

func (p UpdateServiceParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.ID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "id", "must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(p.CategoryID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "category_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if p.DurationMinutes <= 0 {
		errs = sharederrors.AppendValidationError(errs, "duration_minutes", "must be greater than 0")
	}
	if p.Price < 0 {
		errs = sharederrors.AppendValidationError(errs, "price", "must be greater than or equal to 0")
	}
	if !isValidServiceStatus(p.Status) {
		errs = sharederrors.AppendValidationError(errs, "status", "must be 'active' or 'inactive'")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type CreateStaffParams struct {
	SalonID        string
	Name           string
	PhoneNumber    string
	Email          *string
	Role           *string
	Specialization *string
	Photo          *string
	Status         model.StaffStatus
	Shifts         map[string]any
}

func (p CreateStaffParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if strings.TrimSpace(p.PhoneNumber) == "" {
		errs = sharederrors.AppendValidationError(errs, "phone_number", "is required")
	}
	if p.Email != nil && strings.TrimSpace(*p.Email) == "" {
		errs = sharederrors.AppendValidationError(errs, "email", "must not be empty if provided")
	}
	if !isValidStaffStatus(p.Status) {
		errs = sharederrors.AppendValidationError(errs, "status", "must be 'active' or 'inactive'")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type UpdateStaffParams struct {
	ID             string
	SalonID        string
	Name           string
	PhoneNumber    string
	Email          *string
	Role           *string
	Specialization *string
	Photo          *string
	Status         model.StaffStatus
	Shifts         map[string]any
}

func (p UpdateStaffParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if _, err := uuid.Parse(strings.TrimSpace(p.ID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "id", "must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(p.SalonID)); err != nil {
		errs = sharederrors.AppendValidationError(errs, "salon_id", "must be a valid UUID")
	}
	if strings.TrimSpace(p.Name) == "" {
		errs = sharederrors.AppendValidationError(errs, "name", "is required")
	}
	if strings.TrimSpace(p.PhoneNumber) == "" {
		errs = sharederrors.AppendValidationError(errs, "phone_number", "is required")
	}
	if p.Email != nil && strings.TrimSpace(*p.Email) == "" {
		errs = sharederrors.AppendValidationError(errs, "email", "must not be empty if provided")
	}
	if !isValidStaffStatus(p.Status) {
		errs = sharederrors.AppendValidationError(errs, "status", "must be 'active' or 'inactive'")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type RequestStaffOTPParams struct {
	PhoneNumber string
}

func (p RequestStaffOTPParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if strings.TrimSpace(p.PhoneNumber) == "" {
		errs = sharederrors.AppendValidationError(errs, "phone_number", "is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type AuthenticateStaffParams struct {
	PhoneNumber string
	OTP         string
}

func (p AuthenticateStaffParams) Validate() error {
	var errs sharederrors.ValidationErrors
	if strings.TrimSpace(p.PhoneNumber) == "" {
		errs = sharederrors.AppendValidationError(errs, "phone_number", "is required")
	}
	if strings.TrimSpace(p.OTP) == "" {
		errs = sharederrors.AppendValidationError(errs, "otp", "is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func isValidServiceStatus(status model.ServiceStatus) bool {
	switch status {
	case model.ServiceStatusActive, model.ServiceStatusInactive:
		return true
	default:
		return false
	}
}

func isValidStaffStatus(status model.StaffStatus) bool {
	switch status {
	case model.StaffStatusActive, model.StaffStatusInactive:
		return true
	default:
		return false
	}
}
