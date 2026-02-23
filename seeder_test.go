package geeder

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

//go:embed testdata/001_users.sql testdata/002_products.sql
var testSeeds embed.FS

//go:embed testdata/bad/*.sql
var badSeeds embed.FS

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func setupTables(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL)")
	require.NoError(t, err)
}

func TestNew(t *testing.T) {
	db := openTestDB(t)
	fsys, _ := fs.Sub(testSeeds, "testdata")
	s := New(db, fsys)
	assert.NotNil(t, s)
}

func TestRun(t *testing.T) {
	tests := map[string]struct {
		fsys       fs.FS
		setupDB    func(t *testing.T, db *sql.DB)
		wantNames  []string
		wantErr    bool
		wantUsers  int
		wantProdux int
	}{
		"applies all sql files in alphabetical order": {
			fsys: func() fs.FS {
				f, _ := fs.Sub(testSeeds, "testdata")
				return f
			}(),
			setupDB:    setupTables,
			wantNames:  []string{"001_users.sql", "002_products.sql"},
			wantUsers:  1,
			wantProdux: 1,
		},
		"empty fs returns no seeds": {
			fsys:    fstest.MapFS{},
			setupDB: func(t *testing.T, db *sql.DB) {},
		},
		"bad sql rolls back all seeds": {
			fsys: func() fs.FS {
				f, _ := fs.Sub(badSeeds, "testdata/bad")
				return f
			}(),
			setupDB: func(t *testing.T, db *sql.DB) {
				ctx := context.Background()
				_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
				require.NoError(t, err)
			},
			wantErr:   true,
			wantUsers: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db := openTestDB(t)
			ctx := context.Background()
			tt.setupDB(t, db)

			seeds, err := New(db, tt.fsys).Run(ctx)

			if tt.wantErr {
				assert.Error(t, err)

				var count int
				require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count))
				assert.Equal(t, tt.wantUsers, count, "no data after rollback")
				return
			}

			require.NoError(t, err)
			require.Len(t, seeds, len(tt.wantNames))
			for i, wantName := range tt.wantNames {
				assert.Equal(t, wantName, seeds[i].Name, "seed[%d].Name", i)
			}

			if tt.wantUsers > 0 {
				var count int
				require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count))
				assert.Equal(t, tt.wantUsers, count)
			}
			if tt.wantProdux > 0 {
				var count int
				require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&count))
				assert.Equal(t, tt.wantProdux, count)
			}
		})
	}
}

func TestRun_ExecutesEveryTime(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	setupTables(t, db)

	fsys, _ := fs.Sub(testSeeds, "testdata")
	s := New(db, fsys)

	_, err := s.Run(ctx)
	require.NoError(t, err)

	_, err = s.Run(ctx)
	require.NoError(t, err)

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count))
	assert.Equal(t, 2, count, "seeds run every time, no tracking")
}
