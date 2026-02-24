package cli_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/jian-hua-he/geeder/cli"
)

func TestRun(t *testing.T) {
	t.Run("missing dir returns error", func(t *testing.T) {
		t.Setenv("GEEDER_DRIVER", "sqlite")
		t.Setenv("GEEDER_DSN", "foo")

		err := cli.Run(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "-dir flag is required")
	})

	t.Run("missing driver returns error", func(t *testing.T) {
		t.Setenv("GEEDER_DIR", "seeds")
		t.Setenv("GEEDER_DSN", "foo")

		err := cli.Run(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "-driver flag is required")
	})

	t.Run("missing DSN returns error", func(t *testing.T) {
		t.Setenv("GEEDER_DIR", "seeds")
		t.Setenv("GEEDER_DRIVER", "sqlite")

		err := cli.Run(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "-dsn flag is required")
	})

	t.Run("nonexistent dir returns error", func(t *testing.T) {
		t.Setenv("GEEDER_DIR", "/nonexistent/path")
		t.Setenv("GEEDER_DRIVER", "sqlite")
		t.Setenv("GEEDER_DSN", ":memory:")

		err := cli.Run(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "open seed directory")
	})

	t.Run("executes seed files", func(t *testing.T) {
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

		t.Setenv("GEEDER_DIR", seedDir)
		t.Setenv("GEEDER_DRIVER", "sqlite")
		t.Setenv("GEEDER_DSN", tmpDB)

		err = cli.Run(nil)
		require.NoError(t, err)

		db, err = sql.Open("sqlite", tmpDB)
		require.NoError(t, err)
		defer db.Close()

		var count int
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count))
		assert.Equal(t, 1, count)

		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count))
		assert.Equal(t, 1, count)
	})

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

		err = cli.Run(nil)
		require.NoError(t, err)
	})
}
