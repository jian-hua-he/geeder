package geeder

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
)

type mainConfig struct {
	Dir    string
	Driver string
	DSN    string
}

func runMain(cfg mainConfig) error {
	if cfg.Dir == "" {
		return fmt.Errorf("geeder: -dir flag is required")
	}
	if cfg.Driver == "" {
		return fmt.Errorf("geeder: -driver flag is required")
	}
	if cfg.DSN == "" {
		return fmt.Errorf("geeder: -dsn flag is required")
	}

	info, err := os.Stat(cfg.Dir)
	if err != nil {
		return fmt.Errorf("geeder: open seed directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("geeder: %s is not a directory", cfg.Dir)
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("geeder: open database: %w", err)
	}
	defer db.Close()

	seeds, err := New(db, os.DirFS(cfg.Dir)).Run(context.Background())
	if err != nil {
		return err
	}

	for _, s := range seeds {
		fmt.Fprintf(os.Stdout, "applied: %s\n", s.Name)
	}
	return nil
}

// Main is a CLI entry point that reads .sql files from a directory and executes them.
// It parses -dir, -driver, and -dsn flags from the command line.
func Main() {
	dir := flag.String("dir", "", "directory containing .sql seed files")
	driver := flag.String("driver", "", "database driver name (e.g. sqlite, postgres, mysql)")
	dsn := flag.String("dsn", "", "data source name / connection string")
	flag.Parse()

	cfg := mainConfig{
		Dir:    *dir,
		Driver: *driver,
		DSN:    *dsn,
	}

	if err := runMain(cfg); err != nil {
		log.Fatal(err)
	}
}
