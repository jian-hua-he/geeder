// Example: building a CLI seeder binary with embedded SQL files.
//
// Usage:
//
//	go run main.go -driver sqlite -dsn ./example.db
package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/jian-hua-he/geeder"
	_ "modernc.org/sqlite"
)

//go:embed seeds/*.sql
var seedFiles embed.FS

func main() {
	seedFS, err := fs.Sub(seedFiles, "seeds")
	if err != nil {
		log.Fatal(err)
	}
	geeder.Main(seedFS)
}
