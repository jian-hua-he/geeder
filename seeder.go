package geeder

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
)

// Seed represents a single seed file.
type Seed struct {
	Name string // filename, e.g. "001_create_users.sql"
	SQL  string // file contents
}

// Seeder manages seed execution against a database.
type Seeder struct {
	db   *sql.DB
	fsys fs.FS
}

// New creates a Seeder that reads .sql files from the given fs.FS.
// Files are sorted alphabetically by name.
func New(db *sql.DB, fsys fs.FS) *Seeder {
	return &Seeder{db: db, fsys: fsys}
}

// Run executes all .sql files found in the FS.
// All seeds run in a single transaction. If any seed fails, the entire batch is rolled back.
func (s *Seeder) Run(ctx context.Context) ([]Seed, error) {
	seeds, err := s.loadSeeds()
	if err != nil {
		return nil, fmt.Errorf("geeder: load seeds: %w", err)
	}

	if len(seeds) == 0 {
		return nil, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("geeder: begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, seed := range seeds {
		if _, err := tx.ExecContext(ctx, seed.SQL); err != nil {
			return nil, fmt.Errorf("geeder: execute seed %q: %w", seed.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("geeder: commit: %w", err)
	}

	return seeds, nil
}

func (s *Seeder) loadSeeds() ([]Seed, error) {
	matches, err := fs.Glob(s.fsys, "*.sql")
	if err != nil {
		return nil, err
	}

	sort.Strings(matches)

	seeds := make([]Seed, 0, len(matches))
	for _, name := range matches {
		data, err := fs.ReadFile(s.fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", name, err)
		}
		seeds = append(seeds, Seed{Name: name, SQL: string(data)})
	}

	return seeds, nil
}
