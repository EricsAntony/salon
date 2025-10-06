package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GatewayConfig represents payment gateway configuration
type GatewayConfig struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Gateway     string    `json:"gateway" db:"gateway"`
	IsEnabled   bool      `json:"is_enabled" db:"is_enabled"`
	Priority    int       `json:"priority" db:"priority"`
	Config      string    `json:"config" db:"config"` // JSON configuration
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GatewayRepository defines the interface for gateway configuration operations
type GatewayRepository interface {
	GetEnabledGateways(ctx context.Context) ([]*GatewayConfig, error)
	GetGatewayConfig(ctx context.Context, gateway string) (*GatewayConfig, error)
	UpdateGatewayConfig(ctx context.Context, config *GatewayConfig) error
	CreateGatewayConfig(ctx context.Context, config *GatewayConfig) error
}

type gatewayRepository struct {
	db *sql.DB
}

// NewGatewayRepository creates a new gateway repository
func NewGatewayRepository(db *sql.DB) GatewayRepository {
	return &gatewayRepository{db: db}
}

// GetEnabledGateways retrieves all enabled payment gateways ordered by priority
func (r *gatewayRepository) GetEnabledGateways(ctx context.Context) ([]*GatewayConfig, error) {
	query := `
		SELECT id, gateway, is_enabled, priority, config, created_at, updated_at
		FROM gateway_configs 
		WHERE is_enabled = true 
		ORDER BY priority ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled gateways: %w", err)
	}
	defer rows.Close()

	var configs []*GatewayConfig
	for rows.Next() {
		config := &GatewayConfig{}
		err := rows.Scan(
			&config.ID, &config.Gateway, &config.IsEnabled, &config.Priority,
			&config.Config, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gateway config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// GetGatewayConfig retrieves configuration for a specific gateway
func (r *gatewayRepository) GetGatewayConfig(ctx context.Context, gateway string) (*GatewayConfig, error) {
	query := `
		SELECT id, gateway, is_enabled, priority, config, created_at, updated_at
		FROM gateway_configs WHERE gateway = $1`

	config := &GatewayConfig{}
	err := r.db.QueryRowContext(ctx, query, gateway).Scan(
		&config.ID, &config.Gateway, &config.IsEnabled, &config.Priority,
		&config.Config, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("gateway config not found")
		}
		return nil, fmt.Errorf("failed to get gateway config: %w", err)
	}

	return config, nil
}

// UpdateGatewayConfig updates gateway configuration
func (r *gatewayRepository) UpdateGatewayConfig(ctx context.Context, config *GatewayConfig) error {
	query := `
		UPDATE gateway_configs SET
			is_enabled = $2, priority = $3, config = $4, updated_at = $5
		WHERE id = $1`

	config.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		config.ID, config.IsEnabled, config.Priority, config.Config, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update gateway config: %w", err)
	}

	return nil
}

// CreateGatewayConfig creates a new gateway configuration
func (r *gatewayRepository) CreateGatewayConfig(ctx context.Context, config *GatewayConfig) error {
	query := `
		INSERT INTO gateway_configs (id, gateway, is_enabled, priority, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		config.ID, config.Gateway, config.IsEnabled, config.Priority,
		config.Config, config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create gateway config: %w", err)
	}

	return nil
}
