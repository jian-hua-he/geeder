package seeder_test

import (
	"database/sql"
	"embed"
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

	testCases := map[string]struct {
		fsys      fs.FS
		wantErr   bool
		wantSeeds []seeder.Seed
		wantCount int
	}{
		"applies seeds in alphabetical order": {
			fsys: good,
			wantSeeds: []seeder.Seed{
				{Name: "001_users.sql"},
				{Name: "002_users.sql"},
			},
			wantCount: 2,
		},
		"empty filesystem returns no seeds": {
			fsys: fstest.MapFS{},
		},
		"rolls back all seeds on bad SQL": {
			fsys:    bad,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
			require.NoError(t, err)

			s := seeder.New(db, tc.fsys)
			seeds, err := s.Run(t.Context())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tc.wantSeeds, seeds)

			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
			require.NoError(t, err)

			assert.EqualValues(t, tc.wantCount, count)
		})
	}
}

func TestSeeder_Run_ExecutesEveryTime(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
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
	assert.Equal(t, 4, count, "seeds run every invocation, no tracking")
}
