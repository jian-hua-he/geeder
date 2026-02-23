package geeder

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestEnsureTrackingTable(t *testing.T) {
	tests := map[string]struct {
		callCount int
	}{
		"creates table": {
			callCount: 1,
		},
		"idempotent on second call": {
			callCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db := openTestDB(t)
			ctx := context.Background()

			var err error
			for i := 0; i < tt.callCount; i++ {
				err = EnsureTrackingTable(ctx, db)
			}
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, "SELECT name, executed_at FROM geeder_seeds LIMIT 1")
			assert.NoError(t, err, "tracking table should exist")
		})
	}
}

func TestMarkExecuted_And_ExecutedSeeds(t *testing.T) {
	tests := map[string]struct {
		markSeeds []string
		wantIn    []string
		wantNotIn []string
	}{
		"initially empty": {
			markSeeds: nil,
			wantIn:    nil,
			wantNotIn: []string{"anything"},
		},
		"mark one seed": {
			markSeeds: []string{"seed_alpha"},
			wantIn:    []string{"seed_alpha"},
			wantNotIn: []string{"seed_beta"},
		},
		"mark multiple seeds": {
			markSeeds: []string{"seed_1", "seed_2", "seed_3"},
			wantIn:    []string{"seed_1", "seed_2", "seed_3"},
			wantNotIn: []string{"seed_4"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db := openTestDB(t)
			ctx := context.Background()
			require.NoError(t, EnsureTrackingTable(ctx, db))

			for _, s := range tt.markSeeds {
				require.NoError(t, MarkExecuted(ctx, db, s))
			}

			executed, err := ExecutedSeeds(ctx, db)
			require.NoError(t, err)

			for _, s := range tt.wantIn {
				assert.True(t, executed[s], "expected %q in executed set", s)
			}
			for _, s := range tt.wantNotIn {
				assert.False(t, executed[s], "expected %q not in executed set", s)
			}
		})
	}
}

func TestMarkExecuted_DuplicateReturnsError(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	require.NoError(t, EnsureTrackingTable(ctx, db))

	require.NoError(t, MarkExecuted(ctx, db, "dup_seed"))
	err := MarkExecuted(ctx, db, "dup_seed")
	assert.Error(t, err, "marking same seed twice should error")
}
