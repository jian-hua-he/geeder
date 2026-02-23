# geeder

A lightweight Go package for seeding databases with raw SQL. Seeds are registered programmatically in Go code, executed transactionally, and tracked in a `geeder_seeds` table so they never run twice.

Works with any `database/sql`-compatible driver (PostgreSQL, MySQL, SQLite, etc.).

## Installation

```bash
go get github.com/jian-hua-he/geeder
```

## Quick Start

### As a library

```go
package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/jian-hua-he/geeder"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	geeder.Register("001_create_users_table", `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user'
		)
	`)
	geeder.Register("002_seed_admin", "INSERT INTO users (name, role) VALUES ('admin', 'admin')")
	geeder.Register("003_seed_default_users", `
		INSERT INTO users (name) VALUES ('alice');
		INSERT INTO users (name) VALUES ('bob');
	`)

	result, err := geeder.New(db).Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range result.Applied {
		log.Printf("applied: %s", name)
	}
	for _, name := range result.Skipped {
		log.Printf("skipped: %s", name)
	}
}
```

### As a CLI

Build a custom binary that imports your seeds, then run it with flags:

```go
package main

import (
	"github.com/jian-hua-he/geeder"
	_ "myapp/seeds" // triggers init() → geeder.Register() calls
)

func main() {
	geeder.Main()
}
```

```bash
# Run seeds
./myseeder -driver sqlite -dsn ./app.db

# Check status
./myseeder -driver sqlite -dsn ./app.db -status
```

## API

### Register seeds

```go
// Register a seed with a unique name and raw SQL.
// Panics if name or sql is empty, or if the name is already registered.
geeder.Register(name string, sql string)

// Get all registered seeds (returns a copy in registration order).
geeder.Seeds() []geeder.Seed
```

Seeds are executed in **registration order**, giving you explicit control over dependencies.

### Execute seeds

```go
// Create a seeder with a database connection.
s := geeder.New(db)

// Run all registered seeds that haven't been executed yet.
result, err := s.Run(ctx)

// Or run a specific list of seeds (ignores the global registry).
result, err := s.RunSeeds(ctx, []geeder.Seed{
    {Name: "custom_seed", SQL: "INSERT INTO ..."},
})
```

All pending seeds run in a **single transaction**. If any seed fails, the entire batch is rolled back and nothing is tracked.

### Check status

```go
records, err := s.Status(ctx)
for _, r := range records {
    fmt.Printf("%s executed at %s\n", r.Name, r.ExecutedAt)
}
```

### Types

```go
type Seed struct {
    Name string
    SQL  string
}

type Result struct {
    Applied []string // seeds executed in this run
    Skipped []string // seeds already executed previously
}

type SeedRecord struct {
    Name       string
    ExecutedAt time.Time
}
```

## How Tracking Works

Geeder automatically creates a `geeder_seeds` table in your database:

```sql
CREATE TABLE IF NOT EXISTS geeder_seeds (
    name VARCHAR(255) PRIMARY KEY,
    executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)
```

Each time a seed is applied, its name is recorded in this table within the same transaction as the seed SQL. On subsequent runs, already-executed seeds are skipped.

## Examples

See the [`examples/`](examples/) directory for runnable examples:

- **[library](examples/library/)** — Use geeder as a Go library in your application
- **[cli](examples/cli/)** — Build a CLI seeder binary with `geeder.Main()`

## CLI Flags

| Flag | Description | Example |
|---|---|---|
| `-driver` | Database driver name | `sqlite`, `postgres`, `mysql` |
| `-dsn` | Data source name / connection string | `./app.db`, `postgres://user:pass@localhost/db` |
| `-status` | Show executed seeds instead of running | (boolean flag) |

## License

MIT
