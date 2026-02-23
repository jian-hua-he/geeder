# geeder

A lightweight Go tool for seeding databases with raw SQL files. Place your `.sql` files in a directory and geeder executes them in alphabetical order within a single transaction.

Works with any `database/sql`-compatible driver (PostgreSQL, MySQL, SQLite, etc.).

## Installation

### CLI (no Go required)

```bash
go install github.com/jian-hua-he/geeder/cmd/geeder@latest
```

Then run it against any directory of `.sql` files:

```bash
geeder -dir ./seeds -driver postgres -dsn "postgres://user:pass@localhost/mydb"
```

### Library

```bash
go get github.com/jian-hua-he/geeder
```

## Quick Start

### 1. Create your seed files

```
seeds/
├── 001_seed_users.sql
└── 002_seed_products.sql
```

```sql
-- seeds/001_seed_users.sql
INSERT OR IGNORE INTO users (id, name, role) VALUES (1, 'admin', 'admin');
INSERT OR IGNORE INTO users (id, name, role) VALUES (2, 'alice', 'user');
INSERT OR IGNORE INTO users (id, name, role) VALUES (3, 'bob', 'user');
```

```sql
-- seeds/002_seed_products.sql
INSERT OR IGNORE INTO products (id, name, price) VALUES (1, 'Widget', 9.99);
INSERT OR IGNORE INTO products (id, name, price) VALUES (2, 'Gadget', 24.99);
INSERT OR IGNORE INTO products (id, name, price) VALUES (3, 'Gizmo', 49.99);
```

### 2. Run with the CLI

```bash
geeder -dir ./seeds -driver sqlite -dsn ./app.db
```

Output:

```
applied: 001_seed_users.sql
applied: 002_seed_products.sql
```

### 3. Or use as a Go library

```go
package main

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"log"

	"github.com/jian-hua-he/geeder"
	_ "modernc.org/sqlite"
)

//go:embed seeds/*.sql
var seedFiles embed.FS

func main() {
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create schema manually or via a migration tool — geeder is for seeding data.
	// db.Exec(`CREATE TABLE IF NOT EXISTS users (...)`)

	seedFS, _ := fs.Sub(seedFiles, "seeds")
	seeds, err := geeder.New(db, seedFS).Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range seeds {
		log.Printf("applied: %s", s.Name)
	}
}
```

You can also use `os.DirFS` to read from a directory at runtime:

```go
seeds, err := geeder.New(db, os.DirFS("./seeds")).Run(ctx)
```

## API

### Core

```go
// New creates a Seeder that reads .sql files from the given fs.FS.
// Files are sorted alphabetically by name.
geeder.New(db *sql.DB, fsys fs.FS) *Seeder

// Run executes all .sql files in a single transaction.
// Returns the list of seeds that were applied.
(s *Seeder) Run(ctx context.Context) ([]Seed, error)

// Main is a CLI entry point that parses -dir, -driver, and -dsn flags.
geeder.Main()
```

### Types

```go
type Seed struct {
    Name string // filename, e.g. "001_create_users.sql"
    SQL  string // file contents
}
```

## How It Works

1. Geeder reads all `*.sql` files from the provided `fs.FS` (or directory via CLI)
2. Files are sorted alphabetically — use a naming convention like `001_`, `002_` to control order
3. All SQL is executed in a **single transaction** — if any file fails, the entire batch is rolled back
4. Seeds run every time — write idempotent SQL (e.g. `INSERT OR IGNORE`, `ON CONFLICT DO NOTHING`)

## Examples

See the [`examples/`](examples/) directory:

- **[library](examples/library/)** — Use geeder as a Go library with embedded SQL files

## CLI Flags

| Flag | Description | Example |
|---|---|---|
| `-dir` | Directory containing `.sql` seed files | `./seeds` |
| `-driver` | Database driver name | `sqlite`, `postgres`, `mysql` |
| `-dsn` | Data source name / connection string | `./app.db`, `postgres://user:pass@localhost/db` |

## Environment Variables

Flags take precedence. When a flag is omitted, the corresponding environment variable is used:

| Flag | Environment Variable |
|---|---|
| `-dir` | `GEEDER_DIR` |
| `-driver` | `GEEDER_DRIVER` |
| `-dsn` | `GEEDER_DSN` |

```bash
export GEEDER_DIR=./seeds
export GEEDER_DRIVER=postgres
export GEEDER_DSN="postgres://user:pass@localhost/mydb"
geeder
```

## License

MIT
