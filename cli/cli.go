package cli

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/jian-hua-he/geeder/seeder"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("geeder", flag.ContinueOnError)
	dir := fs.String("dir", "", "directory containing .sql seed files (env: GEEDER_DIR)")
	driver := fs.String("driver", "", "database driver name (env: GEEDER_DRIVER)")
	dsn := fs.String("dsn", "", "data source name / connection string (env: GEEDER_DSN)")
	fs.Parse(args)

	if dir == nil || *dir == "" {
		*dir = os.Getenv("GEEDER_DIR")
	}
	if driver == nil || *driver == "" {
		*driver = os.Getenv("GEEDER_DRIVER")
	}
	if dsn == nil || *dsn == "" {
		*dsn = os.Getenv("GEEDER_DSN")
	}

	cfg := Config{
		Dir:    *dir,
		Driver: *driver,
		DSN:    *dsn,
	}

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

	s := seeder.New(db, os.DirFS(cfg.Dir))
	seeds, err := s.Run(context.Background())
	if err != nil {
		return err
	}

	for _, seed := range seeds {
		fmt.Fprintf(os.Stdout, "applied: %s\n", seed.Name)
	}

	return nil
}
