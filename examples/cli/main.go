// Example: building a CLI seeder binary.
//
// This demonstrates how to build a custom seeder command that imports
// your seeds package and delegates to geeder.Main().
//
// Usage:
//
//	go run main.go -driver sqlite -dsn ./example.db
//	go run main.go -driver sqlite -dsn ./example.db -status
package main

import (
	"github.com/jian-hua-he/geeder"

	// Import your seeds package to trigger init() registrations.
	_ "github.com/jian-hua-he/geeder/examples/cli/seeds"

	// Import the database driver you need.
	_ "modernc.org/sqlite"
)

func main() {
	geeder.Main()
}
