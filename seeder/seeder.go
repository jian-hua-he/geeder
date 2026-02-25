package seeder

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Seed represents an applied seed file.
type Seed struct {
	Name string // filename, e.g. "001_create_users.sql"
}

// loadedSeed is the internal representation with SQL content for execution.
type loadedSeed struct {
	name string
	sql  string
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
	loaded, err := s.loadSeeds()
	if err != nil {
		return nil, fmt.Errorf("geeder: load seeds: %w", err)
	}

	if len(loaded) == 0 {
		return nil, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("geeder: begin transaction: %w", err)
	}
	defer tx.Rollback()

	seeds := make([]Seed, 0, len(loaded))
	for _, l := range loaded {
		if _, err := tx.ExecContext(ctx, l.sql); err != nil {
			return nil, fmt.Errorf("geeder: execute seed %q: %w", l.name, err)
		}
		seeds = append(seeds, Seed{Name: l.name})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("geeder: commit: %w", err)
	}

	return seeds, nil
}

func (s *Seeder) loadSeeds() ([]loadedSeed, error) {
	var paths []string
	err := fs.WalkDir(s.fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".sql") {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(paths, func(i, j int) bool {
		return path.Base(paths[i]) < path.Base(paths[j])
	})

	seeds := make([]loadedSeed, 0, len(paths))
	for _, p := range paths {
		data, err := fs.ReadFile(s.fsys, p)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", p, err)
		}
		seeds = append(seeds, loadedSeed{name: path.Base(p), sql: string(data)})
	}

	return seeds, nil
}
