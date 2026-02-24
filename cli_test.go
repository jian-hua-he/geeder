package geeder

import (
	"database/sql"
	"os"
	"path/filepath"
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
		"missing dir": {
			cfg:     mainConfig{Driver: "sqlite", DSN: "foo"},
			wantErr: "-dir flag is required",
		},
		"missing driver": {
			cfg:     mainConfig{Dir: "seeds", DSN: "foo"},
			wantErr: "-driver flag is required",
		},
		"missing DSN": {
			cfg:     mainConfig{Dir: "seeds", Driver: "sqlite"},
			wantErr: "-dsn flag is required",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := runMain(tt.cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestRunMain_ExecutesSeeds(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"

	db, err := sql.Open("sqlite", tmpDB)
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
	require.NoError(t, err)
	_, err = db.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL)")
	require.NoError(t, err)
	db.Close()

	seedDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(seedDir, "001_users.sql"),
		[]byte("INSERT INTO users (name, role) VALUES ('alice', 'admin');"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(seedDir, "002_products.sql"),
		[]byte("INSERT INTO products (name, price) VALUES ('widget', 9.99);"),
		0644,
	))

	err = runMain(mainConfig{Dir: seedDir, Driver: "sqlite", DSN: tmpDB})
	require.NoError(t, err)

	db, err = sql.Open("sqlite", tmpDB)
	require.NoError(t, err)
	defer db.Close()

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count))
	assert.Equal(t, 1, count)

	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count))
	assert.Equal(t, 1, count)
}

func TestRunMain_EnvFallback(t *testing.T) {
	t.Run("env vars are used when flags are empty", func(t *testing.T) {
		tmpDB := t.TempDir() + "/test.db"

		db, err := sql.Open("sqlite", tmpDB)
		require.NoError(t, err)
		_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
		require.NoError(t, err)
		db.Close()

		seedDir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(seedDir, "001_users.sql"),
			[]byte("INSERT INTO users (name, role) VALUES ('alice', 'admin');"),
			0644,
		))

		t.Setenv("GEEDER_DIR", seedDir)
		t.Setenv("GEEDER_DRIVER", "sqlite")
		t.Setenv("GEEDER_DSN", tmpDB)

		err = runMain(mainConfig{})
		require.NoError(t, err)
	})

	t.Run("flags take precedence over env", func(t *testing.T) {
		tmpDB := t.TempDir() + "/test.db"

		db, err := sql.Open("sqlite", tmpDB)
		require.NoError(t, err)
		_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, role TEXT NOT NULL)")
		require.NoError(t, err)
		db.Close()

		seedDir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(seedDir, "001_users.sql"),
			[]byte("INSERT INTO users (name, role) VALUES ('alice', 'admin');"),
			0644,
		))

		t.Setenv("GEEDER_DIR", "/nonexistent")
		t.Setenv("GEEDER_DRIVER", "bad")
		t.Setenv("GEEDER_DSN", "bad")

		err = runMain(mainConfig{Dir: seedDir, Driver: "sqlite", DSN: tmpDB})
		require.NoError(t, err)
	})
}

func TestRunMain_NonexistentDir(t *testing.T) {
	err := runMain(mainConfig{Dir: "/nonexistent/path", Driver: "sqlite", DSN: ":memory:"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open seed directory")
}
