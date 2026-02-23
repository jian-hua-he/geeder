# geeder

A lightweight Go package for seeding databases with raw SQL files. Place your `.sql` files in a directory, embed them with `//go:embed`, and geeder executes them in alphabetical order within a single transaction.

Works with any `database/sql`-compatible driver (PostgreSQL, MySQL, SQLite, etc.).

## Installation

```bash
go get github.com/jian-hua-he/geeder
```

## Quick Start

### 1. Create your seed files

```
seeds/
├── 001_create_users.sql
└── 002_seed_users.sql
```

```sql
-- seeds/001_create_users.sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user'
);
```

```sql
-- seeds/002_seed_users.sql
INSERT OR IGNORE INTO users (id, name, role) VALUES (1, 'admin', 'admin');
INSERT OR IGNORE INTO users (id, name, role) VALUES (2, 'alice', 'user');
```

### 2. Use as a library

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

### 3. Or use as a CLI

```go
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
```

```bash
go run main.go -driver sqlite -dsn ./app.db
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

// Main is a CLI helper that parses -driver and -dsn flags, then runs seeds.
geeder.Main(fsys fs.FS)
```

### Types

```go
type Seed struct {
    Name string // filename, e.g. "001_create_users.sql"
    SQL  string // file contents
}
```

## How It Works

1. Geeder reads all `*.sql` files from the provided `fs.FS`
2. Files are sorted alphabetically — use a naming convention like `001_`, `002_` to control order
3. All SQL is executed in a **single transaction** — if any file fails, the entire batch is rolled back
4. Seeds run every time `Run()` is called — write idempotent SQL (e.g. `CREATE TABLE IF NOT EXISTS`, `INSERT OR IGNORE`)

## Examples

See the [`examples/`](examples/) directory for runnable examples:

- **[library](examples/library/)** — Use geeder as a Go library with embedded SQL files
- **[cli](examples/cli/)** — Build a CLI seeder binary with `geeder.Main()`

## CLI Flags

| Flag | Description | Example |
|---|---|---|
| `-driver` | Database driver name | `sqlite`, `postgres`, `mysql` |
| `-dsn` | Data source name / connection string | `./app.db`, `postgres://user:pass@localhost/db` |

## License

MIT
