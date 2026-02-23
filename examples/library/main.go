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

	"github.com/jian-hua-he/geeder"
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

	// fs.Sub strips the "seeds/" prefix so geeder sees *.sql at root level.
	seedFS, err := fs.Sub(seedFiles, "seeds")
	if err != nil {
		log.Fatal(err)
	}

	seeds, err := geeder.New(db, seedFS).Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Applied seeds ===")
	for _, s := range seeds {
		fmt.Printf("  %s\n", s.Name)
	}

	// Verify data.
	fmt.Println("\n=== Users in database ===")
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
}
