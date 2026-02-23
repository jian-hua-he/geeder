package geeder

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
)

type mainConfig struct {
	Driver string
	DSN    string
}

func runMain(cfg mainConfig, fsys fs.FS) error {
	if cfg.Driver == "" {
		return fmt.Errorf("geeder: -driver flag is required")
	}
	if cfg.DSN == "" {
		return fmt.Errorf("geeder: -dsn flag is required")
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return fmt.Errorf("geeder: open database: %w", err)
	}
	defer db.Close()

	seeds, err := New(db, fsys).Run(context.Background())
	if err != nil {
		return err
	}

	for _, s := range seeds {
		fmt.Fprintf(os.Stdout, "applied: %s\n", s.Name)
	}
	return nil
}

// Main is a convenience entry point for CLI usage.
// The caller's main.go should embed seed files and pass the fs.FS to Main.
func Main(fsys fs.FS) {
	driver := flag.String("driver", "", "database driver name (e.g. sqlite, postgres, mysql)")
	dsn := flag.String("dsn", "", "data source name / connection string")
	flag.Parse()

	cfg := mainConfig{
		Driver: *driver,
		DSN:    *dsn,
	}

	if err := runMain(cfg, fsys); err != nil {
		log.Fatal(err)
	}
}
