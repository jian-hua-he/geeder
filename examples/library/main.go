// Example: using geeder as a Go library.
//
// This example creates an in-memory SQLite database, registers seeds,
// and runs them. Run it with:
//
//	go run main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jianhuahe/geeder"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the target table (in real apps this is done by migrations).
	if _, err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user'
		)
	`); err != nil {
		log.Fatal(err)
	}

	// Register seeds in the order you want them executed.
	geeder.Register("001_admin_user", "INSERT INTO users (name, role) VALUES ('admin', 'admin')")
	geeder.Register("002_default_users", "INSERT INTO users (name) VALUES ('alice')")
	geeder.Register("003_more_users", "INSERT INTO users (name) VALUES ('bob')")

	ctx := context.Background()
	s := geeder.New(db)

	// First run: all seeds are applied.
	fmt.Println("=== First run ===")
	result, err := s.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printResult(result)

	// Second run: all seeds are skipped (idempotent).
	fmt.Println("\n=== Second run ===")
	result, err = s.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	printResult(result)

	// Check status.
	fmt.Println("\n=== Status ===")
	records, err := s.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range records {
		fmt.Printf("  %s  (executed at %s)\n", r.Name, r.ExecutedAt.Format("2006-01-02 15:04:05"))
	}

	// Verify data.
	fmt.Println("\n=== Users in database ===")
	rows, err := db.QueryContext(ctx, "SELECT id, name, role FROM users ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, role string
		if err := rows.Scan(&id, &name, &role); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %d: %s (%s)\n", id, name, role)
	}
}

func printResult(r *geeder.Result) {
	for _, name := range r.Applied {
		fmt.Printf("  applied: %s\n", name)
	}
	for _, name := range r.Skipped {
		fmt.Printf("  skipped: %s\n", name)
	}
	if len(r.Applied) == 0 && len(r.Skipped) == 0 {
		fmt.Println("  (nothing to do)")
	}
}
