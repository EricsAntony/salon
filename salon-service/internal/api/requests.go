package api

import (
	"github.com/EricsAntony/salon/salon-service/internal/service"
	"github.com/EricsAntony/salon/salon-service/internal/model"
)

type salonPayload struct {
	Name               string                 `json:"name"`
	Description        *string                `json:"description,omitempty"`
	Contact            map[string]any         `json:"contact,omitempty"`
	Address            map[string]any         `json:"address,omitempty"`
	GeoLocation        map[string]any         `json:"geo_location,omitempty"`
	Logo               *string                `json:"logo,omitempty"`
	Banner             *string                `json:"banner,omitempty"`
	WorkingHours       map[string]any         `json:"working_hours,omitempty"`
	Holidays           map[string]any         `json:"holidays,omitempty"`
	CancellationPolicy *string                `json:"cancellation_policy,omitempty"`
	PaymentModes       []string               `json:"payment_modes,omitempty"`
	DefaultCurrency    string                 `json:"default_currency"`
	TaxRate            *float64               `json:"tax_rate,omitempty"`
	Settings           map[string]any         `json:"settings,omitempty"`
}

type createSalonRequest salonPayload

type updateSalonRequest salonPayload

type branchPayload struct {
	Name         string                 `json:"name"`
	Address      map[string]any         `json:"address,omitempty"`
	GeoLocation  map[string]any         `json:"geo_location,omitempty"`
	WorkingHours map[string]any         `json:"working_hours,omitempty"`
	Holidays     map[string]any         `json:"holidays,omitempty"`
	Images       []string               `json:"images,omitempty"`
	Contact      map[string]any         `json:"contact,omitempty"`
}

type createBranchRequest branchPayload

type updateBranchRequest branchPayload

type createCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type updateCategoryRequest createCategoryRequest

type createServiceRequest struct {
	CategoryID      string                 `json:"category_id"`
	Name            string                 `json:"name"`
	Description     *string                `json:"description,omitempty"`
	DurationMinutes int                    `json:"duration_minutes"`
	Price           float64                `json:"price"`
	Tags            []string               `json:"tags,omitempty"`
	Status          model.ServiceStatus    `json:"status"`
}

type updateServiceRequest createServiceRequest

type createStaffRequest struct {
	Name           string              `json:"name"`
	Role           *string             `json:"role,omitempty"`
	Specialization *string             `json:"specialization,omitempty"`
	Photo          *string             `json:"photo,omitempty"`
	Status         model.StaffStatus   `json:"status"`
	Shifts         map[string]any      `json:"shifts,omitempty"`
}

type updateStaffRequest createStaffRequest

type setStaffServicesRequest struct {
	ServiceIDs []string `json:"service_ids"`
}

func (p salonPayload) toCreateParams() service.CreateSalonParams {
	tax := 0.0
	if p.TaxRate != nil {
		tax = *p.TaxRate
	}
	return service.CreateSalonParams{
		Name:               p.Name,
		Description:        p.Description,
		Contact:            p.Contact,
		Address:            p.Address,
		GeoLocation:        p.GeoLocation,
		Logo:               p.Logo,
		Banner:             p.Banner,
		WorkingHours:       p.WorkingHours,
		Holidays:           p.Holidays,
		CancellationPolicy: p.CancellationPolicy,
		PaymentModes:       p.PaymentModes,
		DefaultCurrency:    p.DefaultCurrency,
		TaxRate:            tax,
		Settings:           p.Settings,
	}
}

func (p salonPayload) toUpdateParams(id string) service.UpdateSalonParams {
	tax := 0.0
	if p.TaxRate != nil {
		tax = *p.TaxRate
	}
	return service.UpdateSalonParams{
		ID:                 id,
		Name:               p.Name,
		Description:        p.Description,
		Contact:            p.Contact,
		Address:            p.Address,
		GeoLocation:        p.GeoLocation,
		Logo:               p.Logo,
		Banner:             p.Banner,
		WorkingHours:       p.WorkingHours,
		Holidays:           p.Holidays,
		CancellationPolicy: p.CancellationPolicy,
		PaymentModes:       p.PaymentModes,
		DefaultCurrency:    p.DefaultCurrency,
		TaxRate:            tax,
		Settings:           p.Settings,
	}
}

func (p branchPayload) toCreateParams(salonID string) service.CreateBranchParams {
	return service.CreateBranchParams{
		SalonID:      salonID,
		Name:         p.Name,
		Address:      p.Address,
		GeoLocation:  p.GeoLocation,
		WorkingHours: p.WorkingHours,
		Holidays:     p.Holidays,
		Images:       p.Images,
		Contact:      p.Contact,
	}
}

func (p branchPayload) toUpdateParams(salonID, branchID string) service.UpdateBranchParams {
	return service.UpdateBranchParams{
		ID:           branchID,
		SalonID:      salonID,
		Name:         p.Name,
		Address:      p.Address,
		GeoLocation:  p.GeoLocation,
		WorkingHours: p.WorkingHours,
		Holidays:     p.Holidays,
		Images:       p.Images,
		Contact:      p.Contact,
	}
}

func (r createCategoryRequest) toCreateParams(salonID string) service.CreateCategoryParams {
	return service.CreateCategoryParams{
		SalonID:     salonID,
		Name:        r.Name,
		Description: r.Description,
	}
}

func (r updateCategoryRequest) toUpdateParams(salonID, id string) service.UpdateCategoryParams {
	return service.UpdateCategoryParams{
		ID:          id,
		SalonID:     salonID,
		Name:        r.Name,
		Description: r.Description,
	}
}

func (r createServiceRequest) toCreateParams(salonID string) service.CreateServiceParams {
	return service.CreateServiceParams{
		SalonID:         salonID,
		CategoryID:      r.CategoryID,
		Name:            r.Name,
		Description:     r.Description,
		DurationMinutes: r.DurationMinutes,
		Price:           r.Price,
		Tags:            r.Tags,
		Status:          r.Status,
	}
}

func (r updateServiceRequest) toUpdateParams(salonID, id string) service.UpdateServiceParams {
	return service.UpdateServiceParams{
		ID:              id,
		SalonID:         salonID,
		CategoryID:      r.CategoryID,
		Name:            r.Name,
		Description:     r.Description,
		DurationMinutes: r.DurationMinutes,
		Price:           r.Price,
		Tags:            r.Tags,
		Status:          r.Status,
	}
}

func (r createStaffRequest) toCreateParams(salonID string) service.CreateStaffParams {
	return service.CreateStaffParams{
		SalonID:        salonID,
		Name:           r.Name,
		Role:           r.Role,
		Specialization: r.Specialization,
		Photo:          r.Photo,
		Status:         r.Status,
		Shifts:         r.Shifts,
	}
}

func (r updateStaffRequest) toUpdateParams(salonID, id string) service.UpdateStaffParams {
	return service.UpdateStaffParams{
		ID:             id,
		SalonID:        salonID,
		Name:           r.Name,
		Role:           r.Role,
		Specialization: r.Specialization,
		Photo:          r.Photo,
		Status:         r.Status,
		Shifts:         r.Shifts,
	}
}
