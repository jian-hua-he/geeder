package seeder_test

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/jian-hua-he/geeder/seeder"
)

//go:embed testdata/good/*.sql
var goodSeeds embed.FS

//go:embed testdata/bad/*.sql
var badSeeds embed.FS

func TestSeeder_Run(t *testing.T) {
	good, _ := fs.Sub(goodSeeds, "testdata/good")
	bad, _ := fs.Sub(badSeeds, "testdata/bad")

	tests := map[string]struct {
		fsys      fs.FS
		ddl       []string
		wantErr   bool
		wantSeeds []string
		wantRows  map[string]int
	}{
		"applies seeds in alphabetical order": {
			fsys: good,
			ddl: []string{
				"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)",
				"CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL)",
			},
			wantSeeds: []string{"001_users.sql", "002_products.sql"},
			wantRows:  map[string]int{"users": 1, "products": 1},
		},
		"empty filesystem returns no seeds": {
			fsys: fstest.MapFS{},
		},
		"rolls back all seeds on bad SQL": {
			fsys: bad,
			ddl: []string{
				"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)",
			},
			wantErr:  true,
			wantRows: map[string]int{"users": 0},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			for _, stmt := range tt.ddl {
				_, err := db.Exec(stmt)
				require.NoError(t, err)
			}

			seeds, err := seeder.New(db, tt.fsys).Run(t.Context())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, seeds, len(tt.wantSeeds))
				for i, name := range tt.wantSeeds {
					assert.Equal(t, name, seeds[i].Name)
				}
			}

			for table, want := range tt.wantRows {
				var count int
				require.NoError(t, db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count))
				assert.Equal(t, want, count, "row count for %s", table)
			}
		})
	}
}

func TestSeeder_Run_ExecutesEveryTime(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL)")
	require.NoError(t, err)

	fsys, err := fs.Sub(goodSeeds, "testdata/good")
	require.NoError(t, err)

	s := seeder.New(db, fsys)

	_, err = s.Run(t.Context())
	require.NoError(t, err)

	_, err = s.Run(t.Context())
	require.NoError(t, err)

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count))
	assert.Equal(t, 2, count, "seeds run every invocation, no tracking")
}
