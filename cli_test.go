package geeder

import (
	"database/sql"
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
	_, err = db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	db.Close()

	seeds := []Seed{
		{Name: "add_item", SQL: "INSERT INTO items (name) VALUES ('widget')"},
	}

	err = runMain(mainConfig{Driver: "sqlite", DSN: tmpFile}, seeds)
	require.NoError(t, err)

	db, err = sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	defer db.Close()

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count))
	assert.Equal(t, 1, count)
}

func TestRunMain_StatusMode(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"

	db, err := sql.Open("sqlite", tmpFile)
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	db.Close()

	seeds := []Seed{
		{Name: "add_item", SQL: "INSERT INTO items (name) VALUES ('widget')"},
	}

	// Run seeds first
	err = runMain(mainConfig{Driver: "sqlite", DSN: tmpFile}, seeds)
	require.NoError(t, err)

	// Status mode should not error
	err = runMain(mainConfig{Driver: "sqlite", DSN: tmpFile, Status: true}, nil)
	assert.NoError(t, err)
}
