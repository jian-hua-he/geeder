// Example: using geeder as a Go library with embedded SQL files.
//
// Run with:
//
//	go run main.go
package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"

	"github.com/jian-hua-he/geeder/seeder"
	_ "modernc.org/sqlite"
)

//go:embed seeds/*.sql
var seedFiles embed.FS

func main() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create schema manually — geeder is for seeding data, not managing schema.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user'
	)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price REAL NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// fs.Sub strips the "seeds/" prefix so geeder sees *.sql at root level.
	seedFS, err := fs.Sub(seedFiles, "seeds")
	if err != nil {
		log.Fatal(err)
	}

	seeds, err := seeder.New(db, seedFS).Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Applied seeds ===")
	for _, s := range seeds {
		fmt.Printf("  %s\n", s.Name)
	}

	// Verify data.
	fmt.Println("\n=== Users ===")
	rows, err := db.Query("SELECT id, name, role FROM users ORDER BY id")
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

	fmt.Println("\n=== Products ===")
	rows, err = db.Query("SELECT id, name, price FROM products ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		var price float64
		if err := rows.Scan(&id, &name, &price); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %d: %s ($%.2f)\n", id, name, price)
	}
}
