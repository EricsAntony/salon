package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"salon-service/internal/model"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// --- Salon operations ---

func (s *Store) CreateSalon(ctx context.Context, input *model.Salon) (*model.Salon, error) {
	contactJSON, err := mapToJSONB(input.Contact)
	if err != nil {
		return nil, fmt.Errorf("marshal contact: %w", err)
	}
	addressJSON, err := mapToJSONB(input.Address)
	if err != nil {
		return nil, fmt.Errorf("marshal address: %w", err)
	}
	geoJSON, err := mapToJSONB(input.GeoLocation)
	if err != nil {
		return nil, fmt.Errorf("marshal geo_location: %w", err)
	}
	workingJSON, err := mapToJSONB(input.WorkingHours)
	if err != nil {
		return nil, fmt.Errorf("marshal working_hours: %w", err)
	}
	holidayJSON, err := mapToJSONB(input.Holidays)
	if err != nil {
		return nil, fmt.Errorf("marshal holidays: %w", err)
	}
	settingsJSON, err := mapToJSONB(input.Settings)
	if err != nil {
		return nil, fmt.Errorf("marshal settings: %w", err)
	}

	row := s.db.QueryRow(ctx, `
		INSERT INTO salons (
			id, name, description, contact, address, geo_location, logo, banner,
			working_hours, holidays, cancellation_policy, payment_modes,
			default_currency, tax_rate, settings, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, NOW(), NOW()
		) RETURNING
			id, name, description, contact, address, geo_location, logo, banner,
			working_hours, holidays, cancellation_policy, payment_modes,
			default_currency, tax_rate, settings, created_at, updated_at
	`,
		input.ID,
		input.Name,
		input.Description,
		contactJSON,
		addressJSON,
		geoJSON,
		input.Logo,
		input.Banner,
		workingJSON,
		holidayJSON,
		input.CancellationPolicy,
		arrayOrNil(input.PaymentModes),
		input.DefaultCurrency,
		input.TaxRate,
		settingsJSON,
	)
	return scanSalon(row)
}

func (s *Store) GetSalon(ctx context.Context, id string) (*model.Salon, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, name, description, contact, address, geo_location, logo, banner,
			working_hours, holidays, cancellation_policy, payment_modes,
			default_currency, tax_rate, settings, created_at, updated_at
		FROM salons WHERE id = $1
	`, id)
	return scanSalon(row)
}

func (s *Store) ListSalons(ctx context.Context) ([]*model.Salon, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, description, contact, address, geo_location, logo, banner,
			working_hours, holidays, cancellation_policy, payment_modes,
			default_currency, tax_rate, settings, created_at, updated_at
		FROM salons ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salons []*model.Salon
	for rows.Next() {
		salon, err := scanSalon(rows)
		if err != nil {
			return nil, err
		}
		salons = append(salons, salon)
	}
	return salons, rows.Err()
}

func (s *Store) UpdateSalon(ctx context.Context, input *model.Salon) (*model.Salon, error) {
	contactJSON, err := mapToJSONB(input.Contact)
	if err != nil {
		return nil, fmt.Errorf("marshal contact: %w", err)
	}
	addressJSON, err := mapToJSONB(input.Address)
	if err != nil {
		return nil, fmt.Errorf("marshal address: %w", err)
	}
	geoJSON, err := mapToJSONB(input.GeoLocation)
	if err != nil {
		return nil, fmt.Errorf("marshal geo_location: %w", err)
	}
	workingJSON, err := mapToJSONB(input.WorkingHours)
	if err != nil {
		return nil, fmt.Errorf("marshal working_hours: %w", err)
	}
	holidayJSON, err := mapToJSONB(input.Holidays)
	if err != nil {
		return nil, fmt.Errorf("marshal holidays: %w", err)
	}
	settingsJSON, err := mapToJSONB(input.Settings)
	if err != nil {
		return nil, fmt.Errorf("marshal settings: %w", err)
	}

	row := s.db.QueryRow(ctx, `
		UPDATE salons SET
			name = $2,
			description = $3,
			contact = $4,
			address = $5,
			geo_location = $6,
			logo = $7,
			banner = $8,
			working_hours = $9,
			holidays = $10,
			cancellation_policy = $11,
			payment_modes = $12,
			default_currency = $13,
			tax_rate = $14,
			settings = $15,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, contact, address, geo_location, logo, banner,
			working_hours, holidays, cancellation_policy, payment_modes,
			default_currency, tax_rate, settings, created_at, updated_at
	`,
		input.ID,
		input.Name,
		input.Description,
		contactJSON,
		addressJSON,
		geoJSON,
		input.Logo,
		input.Banner,
		workingJSON,
		holidayJSON,
		input.CancellationPolicy,
		arrayOrNil(input.PaymentModes),
		input.DefaultCurrency,
		input.TaxRate,
		settingsJSON,
	)
	return scanSalon(row)
}

func (s *Store) DeleteSalon(ctx context.Context, id string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM salons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Branch operations ---

func (s *Store) CreateBranch(ctx context.Context, input *model.Branch) (*model.Branch, error) {
	addressJSON, err := mapToJSONB(input.Address)
	if err != nil {
		return nil, fmt.Errorf("marshal address: %w", err)
	}
	geoJSON, err := mapToJSONB(input.GeoLocation)
	if err != nil {
		return nil, fmt.Errorf("marshal geo_location: %w", err)
	}
	workingJSON, err := mapToJSONB(input.WorkingHours)
	if err != nil {
		return nil, fmt.Errorf("marshal working_hours: %w", err)
	}
	holidayJSON, err := mapToJSONB(input.Holidays)
	if err != nil {
		return nil, fmt.Errorf("marshal holidays: %w", err)
	}
	contactJSON, err := mapToJSONB(input.Contact)
	if err != nil {
		return nil, fmt.Errorf("marshal contact: %w", err)
	}

	row := s.db.QueryRow(ctx, `
		INSERT INTO branches (
			id, salon_id, name, address, geo_location, working_hours, holidays,
			images, contact, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, NOW(), NOW()
		)
		RETURNING id, salon_id, name, address, geo_location, working_hours, holidays, images, contact, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.Name,
		addressJSON,
		geoJSON,
		workingJSON,
		holidayJSON,
		arrayOrNil(input.Images),
		contactJSON,
	)
	return scanBranch(row)
}

func (s *Store) GetBranch(ctx context.Context, salonID, branchID string) (*model.Branch, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, salon_id, name, address, geo_location, working_hours, holidays, images, contact, created_at, updated_at
		FROM branches WHERE id = $1 AND salon_id = $2
	`, branchID, salonID)
	return scanBranch(row)
}

func (s *Store) ListBranches(ctx context.Context, salonID string) ([]*model.Branch, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, salon_id, name, address, geo_location, working_hours, holidays, images, contact, created_at, updated_at
		FROM branches WHERE salon_id = $1 ORDER BY name
	`, salonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*model.Branch
	for rows.Next() {
		branch, err := scanBranch(rows)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (s *Store) UpdateBranch(ctx context.Context, input *model.Branch) (*model.Branch, error) {
	addressJSON, err := mapToJSONB(input.Address)
	if err != nil {
		return nil, fmt.Errorf("marshal address: %w", err)
	}
	geoJSON, err := mapToJSONB(input.GeoLocation)
	if err != nil {
		return nil, fmt.Errorf("marshal geo_location: %w", err)
	}
	workingJSON, err := mapToJSONB(input.WorkingHours)
	if err != nil {
		return nil, fmt.Errorf("marshal working_hours: %w", err)
	}
	holidayJSON, err := mapToJSONB(input.Holidays)
	if err != nil {
		return nil, fmt.Errorf("marshal holidays: %w", err)
	}
	contactJSON, err := mapToJSONB(input.Contact)
	if err != nil {
		return nil, fmt.Errorf("marshal contact: %w", err)
	}

	row := s.db.QueryRow(ctx, `
		UPDATE branches SET
			name = $3,
			address = $4,
			geo_location = $5,
			working_hours = $6,
			holidays = $7,
			images = $8,
			contact = $9,
			updated_at = NOW()
		WHERE id = $1 AND salon_id = $2
		RETURNING id, salon_id, name, address, geo_location, working_hours, holidays, images, contact, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.Name,
		addressJSON,
		geoJSON,
		workingJSON,
		holidayJSON,
		arrayOrNil(input.Images),
		contactJSON,
	)
	return scanBranch(row)
}

func (s *Store) DeleteBranch(ctx context.Context, salonID, branchID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM branches WHERE id = $1 AND salon_id = $2`, branchID, salonID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Category operations ---

func (s *Store) CreateCategory(ctx context.Context, input *model.Category) (*model.Category, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO categories (
			id, salon_id, name, description, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, NOW(), NOW()
		)
		RETURNING id, salon_id, name, description, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.Name,
		input.Description,
	)
	return scanCategory(row)
}

func (s *Store) GetCategory(ctx context.Context, salonID, categoryID string) (*model.Category, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, salon_id, name, description, created_at, updated_at
		FROM categories WHERE id = $1 AND salon_id = $2
	`, categoryID, salonID)
	return scanCategory(row)
}

func (s *Store) ListCategories(ctx context.Context, salonID string) ([]*model.Category, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, salon_id, name, description, created_at, updated_at
		FROM categories WHERE salon_id = $1 ORDER BY name
	`, salonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		cat, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

func (s *Store) UpdateCategory(ctx context.Context, input *model.Category) (*model.Category, error) {
	row := s.db.QueryRow(ctx, `
		UPDATE categories SET
			name = $3,
			description = $4,
			updated_at = NOW()
		WHERE id = $1 AND salon_id = $2
		RETURNING id, salon_id, name, description, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.Name,
		input.Description,
	)
	return scanCategory(row)
}

func (s *Store) DeleteCategory(ctx context.Context, salonID, categoryID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM categories WHERE id = $1 AND salon_id = $2`, categoryID, salonID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Service operations ---

func (s *Store) CreateService(ctx context.Context, input *model.Service) (*model.Service, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO services (
			id, salon_id, category_id, name, description, duration_minutes, price, tags, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)
		RETURNING id, salon_id, category_id, name, description, duration_minutes, price, tags, status, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.CategoryID,
		input.Name,
		input.Description,
		input.DurationMin,
		input.Price,
		arrayOrNil(input.Tags),
		input.Status,
	)
	return scanService(row)
}

func (s *Store) GetService(ctx context.Context, salonID, serviceID string) (*model.Service, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, salon_id, category_id, name, description, duration_minutes, price, tags, status, created_at, updated_at
		FROM services WHERE id = $1 AND salon_id = $2
	`, serviceID, salonID)
	return scanService(row)
}

func (s *Store) ListServices(ctx context.Context, salonID string, categoryID *string) ([]*model.Service, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if categoryID != nil {
		rows, err = s.db.Query(ctx, `
			SELECT id, salon_id, category_id, name, description, duration_minutes, price, tags, status, created_at, updated_at
			FROM services WHERE salon_id = $1 AND category_id = $2 ORDER BY name
		`, salonID, *categoryID)
	} else {
		rows, err = s.db.Query(ctx, `
			SELECT id, salon_id, category_id, name, description, duration_minutes, price, tags, status, created_at, updated_at
			FROM services WHERE salon_id = $1 ORDER BY name
		`, salonID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*model.Service
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (s *Store) UpdateService(ctx context.Context, input *model.Service) (*model.Service, error) {
	row := s.db.QueryRow(ctx, `
		UPDATE services SET
			category_id = $3,
			name = $4,
			description = $5,
			duration_minutes = $6,
			price = $7,
			tags = $8,
			status = $9,
			updated_at = NOW()
		WHERE id = $1 AND salon_id = $2
		RETURNING id, salon_id, category_id, name, description, duration_minutes, price, tags, status, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.CategoryID,
		input.Name,
		input.Description,
		input.DurationMin,
		input.Price,
		arrayOrNil(input.Tags),
		input.Status,
	)
	return scanService(row)
}

func (s *Store) DeleteService(ctx context.Context, salonID, serviceID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM services WHERE id = $1 AND salon_id = $2`, serviceID, salonID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Staff operations ---

func (s *Store) CreateStaff(ctx context.Context, input *model.Staff) (*model.Staff, error) {
	shiftsJSON, err := mapToJSONB(input.Shifts)
	if err != nil {
		return nil, fmt.Errorf("marshal shifts: %w", err)
	}
	row := s.db.QueryRow(ctx, `
		INSERT INTO staff (
			id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
		)
		RETURNING id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.Name,
		input.PhoneNumber,
		input.Email,
		input.Role,
		input.Specialization,
		input.Photo,
		input.Status,
		shiftsJSON,
	)
	return scanStaff(row)
}

func (s *Store) GetStaff(ctx context.Context, salonID, staffID string) (*model.Staff, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
		FROM staff WHERE id = $1 AND salon_id = $2
	`, staffID, salonID)
	return scanStaff(row)
}

func (s *Store) ListStaff(ctx context.Context, salonID string, status *model.StaffStatus) ([]*model.Staff, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if status != nil {
		rows, err = s.db.Query(ctx, `
			SELECT id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
			FROM staff WHERE salon_id = $1 AND status = $2 ORDER BY name
		`, salonID, *status)
	} else {
		rows, err = s.db.Query(ctx, `
			SELECT id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
			FROM staff WHERE salon_id = $1 ORDER BY name
		`, salonID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staff []*model.Staff
	for rows.Next() {
		st, err := scanStaff(rows)
		if err != nil {
			return nil, err
		}
		staff = append(staff, st)
	}
	return staff, rows.Err()
}

func (s *Store) GetStaffByPhone(ctx context.Context, phone string) (*model.Staff, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
		FROM staff WHERE phone_number = $1
	`, phone)
	st, err := scanStaff(row)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return st, nil
}

func (s *Store) DeleteStaff(ctx context.Context, salonID, staffID string) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM staff WHERE id = $1 AND salon_id = $2`, staffID, salonID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateStaff(ctx context.Context, input *model.Staff) (*model.Staff, error) {
	shiftsJSON, err := mapToJSONB(input.Shifts)
	if err != nil {
		return nil, fmt.Errorf("marshal shifts: %w", err)
	}
	row := s.db.QueryRow(ctx, `
		UPDATE staff SET
			name = $3,
			phone_number = $4,
			email = $5,
			role = $6,
			specialization = $7,
			photo = $8,
			status = $9,
			shifts = $10,
			updated_at = NOW()
		WHERE id = $1 AND salon_id = $2
		RETURNING id, salon_id, name, phone_number, email, role, specialization, photo, status, shifts, created_at, updated_at
	`,
		input.ID,
		input.SalonID,
		input.Name,
		input.PhoneNumber,
		input.Email,
		input.Role,
		input.Specialization,
		input.Photo,
		input.Status,
		shiftsJSON,
	)
	return scanStaff(row)
}

func (s *Store) SetStaffServices(ctx context.Context, salonID, staffID string, serviceIDs []string) (err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	var exists bool
	if err = tx.QueryRow(ctx, `SELECT true FROM staff WHERE id = $1 AND salon_id = $2`, staffID, salonID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if len(serviceIDs) > 0 {
		var count int
		if err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM services WHERE salon_id = $1 AND id = ANY($2)`, salonID, serviceIDs).Scan(&count); err != nil {
			return err
		}
		if count != len(serviceIDs) {
			return fmt.Errorf("one or more services do not belong to salon")
		}
	}

	if _, err = tx.Exec(ctx, `DELETE FROM staff_services WHERE staff_id = $1`, staffID); err != nil {
		return err
	}

	for _, svcID := range serviceIDs {
		if _, err = tx.Exec(ctx, `INSERT INTO staff_services (id, staff_id, service_id) VALUES (gen_random_uuid(), $1, $2)`, staffID, svcID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ListStaffServices(ctx context.Context, staffID string) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT service_id FROM staff_services WHERE staff_id = $1`, staffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serviceIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		serviceIDs = append(serviceIDs, id)
	}
	return serviceIDs, rows.Err()
}

// --- Helpers ---

func scanSalon(row pgx.Row) (*model.Salon, error) {
	var (
		salon              model.Salon
		description        *string
		contactRaw         []byte
		addressRaw         []byte
		geoRaw             []byte
		workingRaw         []byte
		holidaysRaw        []byte
		settingsRaw        []byte
		paymentModes       []string
		cancellationPolicy *string
	)

	err := row.Scan(
		&salon.ID,
		&salon.Name,
		&description,
		&contactRaw,
		&addressRaw,
		&geoRaw,
		&salon.Logo,
		&salon.Banner,
		&workingRaw,
		&holidaysRaw,
		&cancellationPolicy,
		&paymentModes,
		&salon.DefaultCurrency,
		&salon.TaxRate,
		&settingsRaw,
		&salon.CreatedAt,
		&salon.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	salon.Description = description
	salon.CancellationPolicy = cancellationPolicy
	salon.PaymentModes = paymentModes
	salon.Contact = decodeJSONMap(contactRaw)
	salon.Address = decodeJSONMap(addressRaw)
	salon.GeoLocation = decodeJSONMap(geoRaw)
	salon.WorkingHours = decodeJSONMap(workingRaw)
	salon.Holidays = decodeJSONMap(holidaysRaw)
	salon.Settings = decodeJSONMap(settingsRaw)

	return &salon, nil
}

func scanBranch(row pgx.Row) (*model.Branch, error) {
	var (
		branch      model.Branch
		addressRaw  []byte
		geoRaw      []byte
		workingRaw  []byte
		holidaysRaw []byte
		images      []string
		contactRaw  []byte
	)
	err := row.Scan(
		&branch.ID,
		&branch.SalonID,
		&branch.Name,
		&addressRaw,
		&geoRaw,
		&workingRaw,
		&holidaysRaw,
		&images,
		&contactRaw,
		&branch.CreatedAt,
		&branch.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	branch.Images = images
	branch.Address = decodeJSONMap(addressRaw)
	branch.GeoLocation = decodeJSONMap(geoRaw)
	branch.WorkingHours = decodeJSONMap(workingRaw)
	branch.Holidays = decodeJSONMap(holidaysRaw)
	branch.Contact = decodeJSONMap(contactRaw)
	return &branch, nil
}

func scanCategory(row pgx.Row) (*model.Category, error) {
	var cat model.Category
	if err := row.Scan(&cat.ID, &cat.SalonID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func scanService(row pgx.Row) (*model.Service, error) {
	var (
		svc   model.Service
		desc  *string
		tags  []string
	)
	if err := row.Scan(
		&svc.ID,
		&svc.SalonID,
		&svc.CategoryID,
		&svc.Name,
		&desc,
		&svc.DurationMin,
		&svc.Price,
		&tags,
		&svc.Status,
		&svc.CreatedAt,
		&svc.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	svc.Description = desc
	svc.Tags = tags
	return &svc, nil
}

func scanStaff(row pgx.Row) (*model.Staff, error) {
	var (
		st        model.Staff
		phone     string
		email     *string
		role      *string
		special   *string
		photo     *string
		shiftsRaw []byte
	)
	if err := row.Scan(
		&st.ID,
		&st.SalonID,
		&st.Name,
		&phone,
		&email,
		&role,
		&special,
		&photo,
		&st.Status,
		&shiftsRaw,
		&st.CreatedAt,
		&st.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	st.PhoneNumber = phone
	st.Email = email
	st.Role = role
	st.Specialization = special
	st.Photo = photo
	st.Shifts = decodeJSONMap(shiftsRaw)
	return &st, nil
}

func mapToJSONB(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	buf, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func arrayOrNil[T any](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	return in
}

func decodeJSONMap(data []byte) map[string]any {
	if len(data) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// HealthCheck ensures database connectivity.
func (s *Store) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.db.Ping(ctx)
}
