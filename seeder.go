package geeder

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SeedRecord represents a row in the geeder_seeds tracking table.
type SeedRecord struct {
	Name       string
	ExecutedAt time.Time
}

// Result is returned by Run and contains execution details.
type Result struct {
	Applied []string
	Skipped []string
}

// Seeder manages seed execution against a database.
type Seeder struct {
	db *sql.DB
}

// New creates a new Seeder with the given database connection.
func New(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

// Run executes all registered seeds that have not yet been run.
func (s *Seeder) Run(ctx context.Context) (*Result, error) {
	return s.RunSeeds(ctx, Seeds())
}

// RunSeeds executes the given seeds (not the global registry).
func (s *Seeder) RunSeeds(ctx context.Context, seeds []Seed) (*Result, error) {
	if err := EnsureTrackingTable(ctx, s.db); err != nil {
		return nil, fmt.Errorf("geeder: ensure tracking table: %w", err)
	}

	executed, err := ExecutedSeeds(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("geeder: query executed seeds: %w", err)
	}

	result := &Result{}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("geeder: begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, seed := range seeds {
		if executed[seed.Name] {
			result.Skipped = append(result.Skipped, seed.Name)
			continue
		}

		if _, err := tx.ExecContext(ctx, seed.SQL); err != nil {
			return nil, fmt.Errorf("geeder: execute seed %q: %w", seed.Name, err)
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO "+TrackingTable+" (name) VALUES (?)", seed.Name); err != nil {
			return nil, fmt.Errorf("geeder: mark seed %q: %w", seed.Name, err)
		}

		result.Applied = append(result.Applied, seed.Name)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("geeder: commit: %w", err)
	}

	return result, nil
}

// Status returns records of all previously executed seeds.
func (s *Seeder) Status(ctx context.Context) ([]SeedRecord, error) {
	if err := EnsureTrackingTable(ctx, s.db); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT name, executed_at FROM "+TrackingTable+" ORDER BY executed_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SeedRecord
	for rows.Next() {
		var r SeedRecord
		if err := rows.Scan(&r.Name, &r.ExecutedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
