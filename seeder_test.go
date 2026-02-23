package geeder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestNew(t *testing.T) {
	db := openTestDB(t)
	s := New(db)
	assert.NotNil(t, s)
}

func TestRun(t *testing.T) {
	tests := map[string]struct {
		setup       func(t *testing.T)
		seeds       []Seed
		wantApplied []string
		wantSkipped []string
		wantUsers   int
	}{
		"applies all seeds": {
			seeds: []Seed{
				{Name: "seed_1", SQL: "INSERT INTO users (name) VALUES ('alice')"},
				{Name: "seed_2", SQL: "INSERT INTO users (name) VALUES ('bob')"},
			},
			wantApplied: []string{"seed_1", "seed_2"},
			wantSkipped: nil,
			wantUsers:   2,
		},
		"empty registry returns empty result": {
			seeds:       nil,
			wantApplied: nil,
			wantSkipped: nil,
			wantUsers:   0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ResetRegistry()
			db := openTestDB(t)
			ctx := context.Background()

			_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
			require.NoError(t, err)

			for _, s := range tt.seeds {
				Register(s.Name, s.SQL)
			}

			result, err := New(db).Run(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.wantApplied, result.Applied)
			assert.Equal(t, tt.wantSkipped, result.Skipped)

			var count int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count))
			assert.Equal(t, tt.wantUsers, count)
		})
	}
}

func TestRun_Idempotent(t *testing.T) {
	ResetRegistry()
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	require.NoError(t, err)

	Register("seed_1", "INSERT INTO users (name) VALUES ('alice')")
	Register("seed_2", "INSERT INTO users (name) VALUES ('bob')")

	s := New(db)

	result1, err := s.Run(ctx)
	require.NoError(t, err)
	assert.Len(t, result1.Applied, 2)

	result2, err := s.Run(ctx)
	require.NoError(t, err)
	assert.Empty(t, result2.Applied)
	assert.Len(t, result2.Skipped, 2)

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count))
	assert.Equal(t, 2, count, "no duplicate data after second run")
}

func TestRun_PartialApplication(t *testing.T) {
	ResetRegistry()
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	require.NoError(t, err)

	Register("seed_1", "INSERT INTO users (name) VALUES ('alice')")

	s := New(db)
	_, err = s.Run(ctx)
	require.NoError(t, err)

	Register("seed_2", "INSERT INTO users (name) VALUES ('bob')")

	result, err := s.Run(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"seed_2"}, result.Applied)
	assert.Equal(t, []string{"seed_1"}, result.Skipped)
}

func TestRun_BadSQLRollsBack(t *testing.T) {
	ResetRegistry()
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	Register("good_seed", "INSERT INTO users (name) VALUES ('alice')")
	Register("bad_seed", "INSERT INTO nonexistent_table (x) VALUES (1)")

	_, err = New(db).Run(ctx)
	assert.Error(t, err)

	executed, err := ExecutedSeeds(ctx, db)
	require.NoError(t, err)
	assert.False(t, executed["good_seed"], "good_seed should not be tracked after rollback")

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count))
	assert.Equal(t, 0, count, "no data after rollback")
}

func TestRunSeeds_UsesProvidedList(t *testing.T) {
	ResetRegistry()
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	Register("global_seed", "INSERT INTO users (name) VALUES ('global')")

	explicit := []Seed{
		{Name: "explicit_seed", SQL: "INSERT INTO users (name) VALUES ('explicit')"},
	}

	result, err := New(db).RunSeeds(ctx, explicit)
	require.NoError(t, err)

	assert.Equal(t, []string{"explicit_seed"}, result.Applied)

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE name='global'").Scan(&count))
	assert.Equal(t, 0, count, "global seed should not have been applied")
}

func TestStatus(t *testing.T) {
	tests := map[string]struct {
		seeds      []Seed
		wantLen    int
		wantNames  []string
		wantNonZero bool
	}{
		"returns executed seeds after run": {
			seeds: []Seed{
				{Name: "seed_1", SQL: "INSERT INTO users (name) VALUES ('alice')"},
				{Name: "seed_2", SQL: "INSERT INTO users (name) VALUES ('bob')"},
			},
			wantLen:     2,
			wantNames:   []string{"seed_1", "seed_2"},
			wantNonZero: true,
		},
		"empty when nothing run": {
			seeds:   nil,
			wantLen: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ResetRegistry()
			db := openTestDB(t)
			ctx := context.Background()

			_, err := db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
			require.NoError(t, err)

			for _, s := range tt.seeds {
				Register(s.Name, s.SQL)
			}

			s := New(db)
			if len(tt.seeds) > 0 {
				_, err = s.Run(ctx)
				require.NoError(t, err)
			}

			records, err := s.Status(ctx)
			require.NoError(t, err)
			assert.Len(t, records, tt.wantLen)

			if tt.wantNonZero {
				names := make(map[string]bool)
				for _, r := range records {
					names[r.Name] = true
					assert.False(t, r.ExecutedAt.IsZero(), "ExecutedAt should not be zero for %s", r.Name)
				}
				for _, wn := range tt.wantNames {
					assert.True(t, names[wn], "expected %q in status", wn)
				}
			}
		})
	}
}
