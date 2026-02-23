package geeder

import (
	"database/sql"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestRunMain_Validation(t *testing.T) {
	tests := map[string]struct {
		cfg     mainConfig
		wantErr string
	}{
		"missing driver": {
			cfg:     mainConfig{DSN: "foo"},
			wantErr: "-driver flag is required",
		},
		"missing DSN": {
			cfg:     mainConfig{Driver: "sqlite"},
			wantErr: "-dsn flag is required",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := runMain(tt.cfg, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestRunMain_ExecutesSeeds(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	db, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL)")
	require.NoError(t, err)
	db.Close()

	fsys, err := fs.Sub(testSeeds, "testdata")
	require.NoError(t, err)

	err = runMain(mainConfig{Driver: "sqlite", DSN: tmpFile}, fsys)
	require.NoError(t, err)

	db, err = sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	defer db.Close()

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count))
	assert.Equal(t, 1, count)
}
