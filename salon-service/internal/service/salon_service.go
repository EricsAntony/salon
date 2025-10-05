package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"salon-service/internal/model"
	"salon-service/internal/repository"

	sharedauth "github.com/EricsAntony/salon/salon-shared/auth"
	sharedconfig "github.com/EricsAntony/salon/salon-shared/config"
	sharederrors "github.com/EricsAntony/salon/salon-shared/errors"
	sharedvalidation "github.com/EricsAntony/salon/salon-shared/validation"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SalonService interface {
	CreateSalon(ctx context.Context, params CreateSalonParams) (*model.Salon, error)
	GetSalon(ctx context.Context, id string) (*model.Salon, error)
	ListSalons(ctx context.Context) ([]*model.Salon, error)
	UpdateSalon(ctx context.Context, params UpdateSalonParams) (*model.Salon, error)
	DeleteSalon(ctx context.Context, id string) error

	CreateBranch(ctx context.Context, params CreateBranchParams) (*model.Branch, error)
	GetBranch(ctx context.Context, salonID, branchID string) (*model.Branch, error)
	ListBranches(ctx context.Context, salonID string) ([]*model.Branch, error)
	UpdateBranch(ctx context.Context, params UpdateBranchParams) (*model.Branch, error)
	DeleteBranch(ctx context.Context, salonID, branchID string) error

	CreateCategory(ctx context.Context, params CreateCategoryParams) (*model.Category, error)
	ListCategories(ctx context.Context, salonID string) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, params UpdateCategoryParams) (*model.Category, error)
	DeleteCategory(ctx context.Context, salonID, categoryID string) error

	CreateService(ctx context.Context, params CreateServiceParams) (*model.Service, error)
	ListServices(ctx context.Context, salonID string, categoryID *string) ([]*model.Service, error)
	UpdateService(ctx context.Context, params UpdateServiceParams) (*model.Service, error)
	DeleteService(ctx context.Context, salonID, serviceID string) error

	CreateStaff(ctx context.Context, params CreateStaffParams) (*model.Staff, error)
	ListStaff(ctx context.Context, salonID string, status *model.StaffStatus) ([]*model.Staff, error)
	UpdateStaff(ctx context.Context, params UpdateStaffParams) (*model.Staff, error)
	DeleteStaff(ctx context.Context, salonID, staffID string) error
	SetStaffServices(ctx context.Context, salonID, staffID string, serviceIDs []string) error
	ListStaffServices(ctx context.Context, staffID string) ([]string, error)
	RequestStaffOTP(ctx context.Context, params RequestStaffOTPParams) error
	AuthenticateStaff(ctx context.Context, params AuthenticateStaffParams) (*AuthenticateStaffResult, error)
	RefreshStaffSession(ctx context.Context, staffID, refreshToken string) (*AuthenticateStaffResult, error)

	HealthCheck(ctx context.Context) error
}

type salonService struct {
	repo      *repository.Store
	authRepo  repository.StaffAuthRepository
	cfg       *sharedconfig.Config
	jwt       JWTIssuer
	otpExpiry time.Duration
}

type JWTIssuer interface {
	GenerateAccessToken(userID string) (string, time.Time, error)
	GenerateRefreshToken(userID string) (string, time.Time, error)
	GenerateAccessTokenWithType(userID, userType string) (string, time.Time, error)
	GenerateRefreshTokenWithType(userID, userType string) (string, time.Time, error)
}

func New(repo *repository.Store, authRepo repository.StaffAuthRepository, cfg *sharedconfig.Config, jwt JWTIssuer) SalonService {
	return &salonService{
		repo:      repo,
		authRepo:  authRepo,
		cfg:       cfg,
		jwt:       jwt,
		otpExpiry: time.Duration(cfg.OTP.ExpiryMinutes) * time.Minute,
	}
}

func (s *salonService) CreateSalon(ctx context.Context, params CreateSalonParams) (*model.Salon, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	salon := &model.Salon{
		ID:                 uuid.NewString(),
		Name:               params.Name,
		Description:        params.Description,
		Contact:            params.Contact,
		Address:            params.Address,
		GeoLocation:        params.GeoLocation,
		Logo:               params.Logo,
		Banner:             params.Banner,
		WorkingHours:       params.WorkingHours,
		Holidays:           params.Holidays,
		CancellationPolicy: params.CancellationPolicy,
		PaymentModes:       params.PaymentModes,
		DefaultCurrency:    params.DefaultCurrency,
		TaxRate:            params.TaxRate,
		Settings:           params.Settings,
	}
	return s.repo.CreateSalon(ctx, salon)
}

func (s *salonService) GetSalon(ctx context.Context, id string) (*model.Salon, error) {
	if err := validateUUID("salon_id", id); err != nil {
		return nil, err
	}
	return s.repo.GetSalon(ctx, id)
}


func (s *salonService) ListSalons(ctx context.Context) ([]*model.Salon, error) {
	return s.repo.ListSalons(ctx)
}

func (s *salonService) UpdateSalon(ctx context.Context, params UpdateSalonParams) (*model.Salon, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	salon := &model.Salon{
		ID:                 params.ID,
		Name:               params.Name,
		Description:        params.Description,
		Contact:            params.Contact,
		Address:            params.Address,
		GeoLocation:        params.GeoLocation,
		Logo:               params.Logo,
		Banner:             params.Banner,
		WorkingHours:       params.WorkingHours,
		Holidays:           params.Holidays,
		CancellationPolicy: params.CancellationPolicy,
		PaymentModes:       params.PaymentModes,
		DefaultCurrency:    params.DefaultCurrency,
		TaxRate:            params.TaxRate,
		Settings:           params.Settings,
	}
	return s.repo.UpdateSalon(ctx, salon)
}

func (s *salonService) DeleteSalon(ctx context.Context, id string) error {
	if err := validateUUID("salon_id", id); err != nil {
		return err
	}
	return s.repo.DeleteSalon(ctx, id)
}

func (s *salonService) CreateBranch(ctx context.Context, params CreateBranchParams) (*model.Branch, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	branch := &model.Branch{
		ID:           uuid.NewString(),
		SalonID:      params.SalonID,
		Name:         params.Name,
		Address:      params.Address,
		GeoLocation:  params.GeoLocation,
		WorkingHours: params.WorkingHours,
		Holidays:     params.Holidays,
		Images:       params.Images,
		Contact:      params.Contact,
	}
	return s.repo.CreateBranch(ctx, branch)
}

func (s *salonService) GetBranch(ctx context.Context, salonID, branchID string) (*model.Branch, error) {
	if err := validateUUID("salon_id", salonID); err != nil {
		return nil, err
	}
	if err := validateUUID("branch_id", branchID); err != nil {
		return nil, err
	}
	return s.repo.GetBranch(ctx, salonID, branchID)
}

func (s *salonService) ListBranches(ctx context.Context, salonID string) ([]*model.Branch, error) {
	if err := validateUUID("salon_id", salonID); err != nil {
		return nil, err
	}
	return s.repo.ListBranches(ctx, salonID)
}

func (s *salonService) UpdateBranch(ctx context.Context, params UpdateBranchParams) (*model.Branch, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	branch := &model.Branch{
		ID:           params.ID,
		SalonID:      params.SalonID,
		Name:         params.Name,
		Address:      params.Address,
		GeoLocation:  params.GeoLocation,
		WorkingHours: params.WorkingHours,
		Holidays:     params.Holidays,
		Images:       params.Images,
		Contact:      params.Contact,
	}
	return s.repo.UpdateBranch(ctx, branch)
}

func (s *salonService) DeleteBranch(ctx context.Context, salonID, branchID string) error {
	if err := validateUUID("salon_id", salonID); err != nil {
		return err
	}
	if err := validateUUID("branch_id", branchID); err != nil {
		return err
	}
	return s.repo.DeleteBranch(ctx, salonID, branchID)
}

func (s *salonService) CreateCategory(ctx context.Context, params CreateCategoryParams) (*model.Category, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	category := &model.Category{
		ID:          uuid.NewString(),
		SalonID:     params.SalonID,
		Name:        params.Name,
		Description: params.Description,
	}
	return s.repo.CreateCategory(ctx, category)
}

func (s *salonService) ListCategories(ctx context.Context, salonID string) ([]*model.Category, error) {
	if err := validateUUID("salon_id", salonID); err != nil {
		return nil, err
	}
	return s.repo.ListCategories(ctx, salonID)
}

func (s *salonService) UpdateCategory(ctx context.Context, params UpdateCategoryParams) (*model.Category, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	category := &model.Category{
		ID:          params.ID,
		SalonID:     params.SalonID,
		Name:        params.Name,
		Description: params.Description,
	}
	return s.repo.UpdateCategory(ctx, category)
}

func (s *salonService) DeleteCategory(ctx context.Context, salonID, categoryID string) error {
	if err := validateUUID("salon_id", salonID); err != nil {
		return err
	}
	if err := validateUUID("category_id", categoryID); err != nil {
		return err
	}
	return s.repo.DeleteCategory(ctx, salonID, categoryID)
}

func (s *salonService) CreateService(ctx context.Context, params CreateServiceParams) (*model.Service, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	service := &model.Service{
		ID:          uuid.NewString(),
		SalonID:     params.SalonID,
		CategoryID:  params.CategoryID,
		Name:        params.Name,
		Description: params.Description,
		DurationMin: params.DurationMinutes,
		Price:       params.Price,
		Tags:        params.Tags,
		Status:      params.Status,
	}
	return s.repo.CreateService(ctx, service)
}

func (s *salonService) ListServices(ctx context.Context, salonID string, categoryID *string) ([]*model.Service, error) {
	if err := validateUUID("salon_id", salonID); err != nil {
		return nil, err
	}
	if categoryID != nil {
		if err := validateUUID("category_id", *categoryID); err != nil {
			return nil, err
		}
	}
	return s.repo.ListServices(ctx, salonID, categoryID)
}

func (s *salonService) UpdateService(ctx context.Context, params UpdateServiceParams) (*model.Service, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	service := &model.Service{
		ID:          params.ID,
		SalonID:     params.SalonID,
		CategoryID:  params.CategoryID,
		Name:        params.Name,
		Description: params.Description,
		DurationMin: params.DurationMinutes,
		Price:       params.Price,
		Tags:        params.Tags,
		Status:      params.Status,
	}
	return s.repo.UpdateService(ctx, service)
}

func (s *salonService) DeleteService(ctx context.Context, salonID, serviceID string) error {
	if err := validateUUID("salon_id", salonID); err != nil {
		return err
	}
	if err := validateUUID("service_id", serviceID); err != nil {
		return err
	}
	return s.repo.DeleteService(ctx, salonID, serviceID)
}

func (s *salonService) CreateStaff(ctx context.Context, params CreateStaffParams) (*model.Staff, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	// Use shared phone validation and normalization
	normalizedPhone, err := sharedvalidation.ValidatePhone(params.PhoneNumber)
	if err != nil {
		return nil, err
	}
	params.PhoneNumber = normalizedPhone
	staff := &model.Staff{
		ID:             uuid.NewString(),
		SalonID:        params.SalonID,
		Name:           params.Name,
		PhoneNumber:    params.PhoneNumber,
		Email:          params.Email,
		Role:           params.Role,
		Specialization: params.Specialization,
		Photo:          params.Photo,
		Status:         params.Status,
		Shifts:         params.Shifts,
	}
	created, err := s.repo.CreateStaff(ctx, staff)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return nil, sharederrors.NewValidationError("phone_number", "already registered")
		}
		return nil, err
	}
	return created, nil
}

func (s *salonService) ListStaff(ctx context.Context, salonID string, status *model.StaffStatus) ([]*model.Staff, error) {
	if err := validateUUID("salon_id", salonID); err != nil {
		return nil, err
	}
	return s.repo.ListStaff(ctx, salonID, status)
}

func (s *salonService) UpdateStaff(ctx context.Context, params UpdateStaffParams) (*model.Staff, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	// Use shared phone validation and normalization
	normalizedPhone, err := sharedvalidation.ValidatePhone(params.PhoneNumber)
	if err != nil {
		return nil, err
	}
	params.PhoneNumber = normalizedPhone
	staff := &model.Staff{
		ID:             params.ID,
		SalonID:        params.SalonID,
		Name:           params.Name,
		PhoneNumber:    params.PhoneNumber,
		Email:          params.Email,
		Role:           params.Role,
		Specialization: params.Specialization,
		Photo:          params.Photo,
		Status:         params.Status,
		Shifts:         params.Shifts,
	}
	updated, err := s.repo.UpdateStaff(ctx, staff)
	if err != nil {
		if repository.IsUniqueViolation(err) {
			return nil, sharederrors.NewValidationError("phone_number", "already registered")
		}
		return nil, err
	}
	return updated, nil
}

func (s *salonService) DeleteStaff(ctx context.Context, salonID, staffID string) error {
	if err := validateUUID("salon_id", salonID); err != nil {
		return err
	}
	if err := validateUUID("staff_id", staffID); err != nil {
		return err
	}
	return s.repo.DeleteStaff(ctx, salonID, staffID)
}

func (s *salonService) SetStaffServices(ctx context.Context, salonID, staffID string, serviceIDs []string) error {
	if err := validateUUID("salon_id", salonID); err != nil {
		return err
	}
	if err := validateUUID("staff_id", staffID); err != nil {
		return err
	}
	for _, id := range serviceIDs {
		if err := validateUUID("service_id", id); err != nil {
			return err
		}
	}
	return s.repo.SetStaffServices(ctx, salonID, staffID, serviceIDs)
}

func (s *salonService) ListStaffServices(ctx context.Context, staffID string) ([]string, error) {
	if err := validateUUID("staff_id", staffID); err != nil {
		return nil, err
	}
	return s.repo.ListStaffServices(ctx, staffID)
}

func (s *salonService) RequestStaffOTP(ctx context.Context, params RequestStaffOTPParams) error {
	// Use shared phone validation and normalization
	phone, err := sharedvalidation.ValidatePhone(params.PhoneNumber)
	if err != nil {
		return err
	}

	// Check if staff exists with this phone number
	staff, err := s.repo.GetStaffByPhone(ctx, phone)
	if err != nil {
		return err
	}
	if staff == nil {
		return ErrStaffNotFound
	}

	// Generate OTP using shared utility
	otpCode, err := sharedvalidation.GenerateOTP()
	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}

	// Store OTP with expiry
	expiry := time.Now().Add(s.otpExpiry)
	if err := s.authRepo.CreateOTP(ctx, staff.ID, otpCode, expiry); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}

	// Use shared OTP sending utility
	sharedauth.SendOTPViaSMS(phone, otpCode)

	return nil
}

type AuthenticateStaffResult struct {
	Staff        *model.Staff `json:"staff"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

var (
	ErrStaffNotFound       = errors.New("staff not found")
	ErrInvalidOTP          = errors.New("invalid otp")
	ErrOTPExpired          = errors.New("otp expired")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
)

func (s *salonService) AuthenticateStaff(ctx context.Context, params AuthenticateStaffParams) (*AuthenticateStaffResult, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	// Use shared phone validation and normalization
	phone, err := sharedvalidation.ValidatePhone(params.PhoneNumber)
	if err != nil {
		return nil, err
	}

	staff, err := s.repo.GetStaffByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if staff == nil {
		return nil, ErrStaffNotFound
	}

	otp, err := s.authRepo.GetLatestOTP(ctx, phone)
	if err != nil {
		return nil, err
	}
	if otp == nil {
		return nil, ErrInvalidOTP
	}
	if time.Now().After(otp.ExpiresAt) {
		return nil, ErrOTPExpired
	}
	if err := bcrypt.CompareHashAndPassword([]byte(otp.CodeHash), []byte(params.OTP)); err != nil {
		_ = s.authRepo.IncrementOTPAttempts(ctx, otp.ID)
		return nil, ErrInvalidOTP
	}

	return s.issueTokens(ctx, staff)
}

func (s *salonService) RefreshStaffSession(ctx context.Context, staffID, refreshToken string) (*AuthenticateStaffResult, error) {
	if err := validateUUID("staff_id", staffID); err != nil {
		return nil, err
	}

	staff, err := s.repo.GetStaff(ctx, staffID, staffID)
	if err != nil {
		return nil, err
	}
	if staff == nil {
		return nil, ErrStaffNotFound
	}

	valid, err := s.authRepo.IsRefreshTokenValid(ctx, staff.ID, refreshToken, time.Now())
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrRefreshTokenInvalid
	}

	return s.issueTokens(ctx, staff)
}

func (s *salonService) issueTokens(ctx context.Context, staff *model.Staff) (*AuthenticateStaffResult, error) {
	accessToken, _, err := s.jwt.GenerateAccessTokenWithType(staff.ID, sharedauth.UserTypeSalon)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshExpiry, err := s.jwt.GenerateRefreshTokenWithType(staff.ID, sharedauth.UserTypeSalon)
	if err != nil {
		return nil, err
	}
	refreshHash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash refresh token: %w", err)
	}
	if err := s.authRepo.CreateRefreshToken(ctx, staff.ID, string(refreshHash), refreshExpiry); err != nil {
		return nil, err
	}

	return &AuthenticateStaffResult{
		Staff:        staff,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}


func (s *salonService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}

func validateUUID(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return sharederrors.NewValidationError(field, "must not be empty")
	}
	if _, err := uuid.Parse(value); err != nil {
		return sharederrors.NewValidationError(field, "must be a valid UUID")
	}
	return nil
}
