package geeder

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	tests := map[string]struct {
		seeds     []Seed
		wantLen   int
		wantOrder []string
	}{
		"single seed": {
			seeds:     []Seed{{Name: "seed_1", SQL: "INSERT INTO users (name) VALUES ('alice')"}},
			wantLen:   1,
			wantOrder: []string{"seed_1"},
		},
		"multiple seeds preserves registration order": {
			seeds: []Seed{
				{Name: "b_seed", SQL: "SELECT 1"},
				{Name: "a_seed", SQL: "SELECT 2"},
			},
			wantLen:   2,
			wantOrder: []string{"b_seed", "a_seed"},
		},
		"empty registry": {
			seeds:     nil,
			wantLen:   0,
			wantOrder: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ResetRegistry()
			for _, s := range tt.seeds {
				Register(s.Name, s.SQL)
			}

			got := Seeds()
			require.Len(t, got, tt.wantLen)
			for i, wantName := range tt.wantOrder {
				assert.Equal(t, wantName, got[i].Name, "Seeds()[%d].Name", i)
			}
		})
	}
}

func TestRegister_Panics(t *testing.T) {
	tests := map[string]struct {
		setup   func()
		regName string
		regSQL  string
	}{
		"duplicate name": {
			setup:   func() { Register("dup", "SELECT 1") },
			regName: "dup",
			regSQL:  "SELECT 2",
		},
		"empty name": {
			setup:   func() {},
			regName: "",
			regSQL:  "SELECT 1",
		},
		"empty SQL": {
			setup:   func() {},
			regName: "valid_name",
			regSQL:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ResetRegistry()
			tt.setup()

			assert.Panics(t, func() {
				Register(tt.regName, tt.regSQL)
			})
		})
	}
}

func TestSeeds_ReturnsCopy(t *testing.T) {
	ResetRegistry()
	Register("original", "SELECT 1")

	seeds := Seeds()
	seeds[0].Name = "mutated"

	fresh := Seeds()
	assert.Equal(t, "original", fresh[0].Name, "Seeds() should return a copy")
}

func TestResetRegistry(t *testing.T) {
	Register("temp", "SELECT 1")
	ResetRegistry()

	assert.Empty(t, Seeds())
}

func TestRegister_ConcurrentRegistration(t *testing.T) {
	ResetRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			defer func() { recover() }()
			Register(fmt.Sprintf("seed_%d", n), "SELECT 1")
		}(i)
	}
	wg.Wait()

	assert.NotEmpty(t, Seeds())
}
