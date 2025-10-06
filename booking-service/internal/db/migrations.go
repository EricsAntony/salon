package db

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// RunMigrations runs all pending database migrations
func RunMigrations(databaseURL string) error {
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database for migrations: %w", err)
	}
	defer conn.Close(context.Background())

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(conn); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(conn)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; !applied {
			if err := runMigration(conn, file); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", file, err)
			}
			log.Info().Str("migration", file).Msg("Migration applied successfully")
		}
	}

	log.Info().Msg("All migrations completed successfully")
	return nil
}

func createMigrationsTable(conn *pgx.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	_, err := conn.Exec(context.Background(), query)
	return err
}

func getMigrationFiles() ([]string, error) {
	var files []string
	
	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, filepath.Base(path))
		}
		
		return nil
	})
	
	if err != nil {
		// If migrations directory doesn't exist, return empty list
		return []string{}, nil
	}
	
	// Sort files to ensure proper order
	sort.Strings(files)
	return files, nil
}

func getAppliedMigrations(conn *pgx.Conn) (map[string]bool, error) {
	query := "SELECT version FROM schema_migrations"
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func runMigration(conn *pgx.Conn, filename string) error {
	// Read migration file
	content, err := fs.ReadFile(nil, filepath.Join("migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Start transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Execute migration
	if _, err := tx.Exec(context.Background(), string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	if _, err := tx.Exec(context.Background(), "INSERT INTO schema_migrations (version) VALUES ($1)", filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}
