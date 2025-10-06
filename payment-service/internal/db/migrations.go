package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

// RunMigrations runs database migrations from the specified directory
func RunMigrations(db *sql.DB, migrationsPath string) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, file := range migrationFiles {
		migrationName := strings.TrimSuffix(file, ".sql")
		
		if appliedMigrations[migrationName] {
			log.Debug().Str("migration", migrationName).Msg("Migration already applied")
			continue
		}

		log.Info().Str("migration", migrationName).Msg("Applying migration")
		
		if err := applyMigration(db, migrationsPath, file, migrationName); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migrationName, err)
		}
		
		log.Info().Str("migration", migrationName).Msg("Migration applied successfully")
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

	_, err := db.Exec(query)
	return err
}

// getMigrationFiles returns sorted list of migration files
func getMigrationFiles(migrationsPath string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	// Sort files to ensure consistent order
	sort.Strings(files)
	return files, nil
}

// getAppliedMigrations returns a map of applied migration names
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	query := "SELECT migration_name FROM schema_migrations"
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var migrationName string
		if err := rows.Scan(&migrationName); err != nil {
			return nil, err
		}
		applied[migrationName] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration file
func applyMigration(db *sql.DB, migrationsPath, fileName, migrationName string) error {
	// Read migration file
	filePath := filepath.Join(migrationsPath, fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	if _, err := tx.Exec("INSERT INTO schema_migrations (migration_name) VALUES ($1)", migrationName); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}
