package geeder

import (
	"context"
	"database/sql"
)

const TrackingTable = "geeder_seeds"

// EnsureTrackingTable creates the tracking table if it does not exist.
func EnsureTrackingTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS `+TrackingTable+` (
			name VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// MarkExecuted records that a seed has been executed.
func MarkExecuted(ctx context.Context, db *sql.DB, name string) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO "+TrackingTable+" (name) VALUES (?)", name)
	return err
}

// ExecutedSeeds returns the set of seed names already executed.
func ExecutedSeeds(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM "+TrackingTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result[name] = true
	}
	return result, rows.Err()
}
