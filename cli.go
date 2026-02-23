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
	Driver string
	DSN    string
	Status bool
}

func runMain(cfg mainConfig, seeds []Seed) error {
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

	ctx := context.Background()
	s := New(db)

	if cfg.Status {
		records, err := s.Status(ctx)
		if err != nil {
			return err
		}
		for _, r := range records {
			fmt.Fprintf(os.Stdout, "%s\t%s\n", r.Name, r.ExecutedAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	}

	result, err := s.RunSeeds(ctx, seeds)
	if err != nil {
		return err
	}

	for _, name := range result.Applied {
		fmt.Fprintf(os.Stdout, "applied: %s\n", name)
	}
	for _, name := range result.Skipped {
		fmt.Fprintf(os.Stdout, "skipped: %s\n", name)
	}
	return nil
}

// Main is a convenience entry point for CLI usage.
// The caller's main.go should import their seeds package (triggering Register calls)
// and then call geeder.Main().
func Main() {
	driver := flag.String("driver", "", "database driver name (e.g. sqlite, postgres, mysql)")
	dsn := flag.String("dsn", "", "data source name / connection string")
	status := flag.Bool("status", false, "show seed execution status")
	flag.Parse()

	cfg := mainConfig{
		Driver: *driver,
		DSN:    *dsn,
		Status: *status,
	}

	if err := runMain(cfg, Seeds()); err != nil {
		log.Fatal(err)
	}
}
