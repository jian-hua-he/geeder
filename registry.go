package geeder

import (
	"fmt"
	"sync"
)

var (
	mu       sync.Mutex
	registry []Seed
	nameSet  = make(map[string]bool)
)

// Register adds a seed to the global registry.
// Panics if name or sql is empty, or if a seed with the same name is already registered.
func Register(name string, sql string) {
	mu.Lock()
	defer mu.Unlock()

	if name == "" {
		panic("geeder: seed name must not be empty")
	}
	if sql == "" {
		panic("geeder: seed SQL must not be empty")
	}
	if nameSet[name] {
		panic(fmt.Sprintf("geeder: seed %q already registered", name))
	}

	nameSet[name] = true
	registry = append(registry, Seed{Name: name, SQL: sql})
}

// Seeds returns a copy of all registered seeds in registration order.
func Seeds() []Seed {
	mu.Lock()
	defer mu.Unlock()

	cp := make([]Seed, len(registry))
	copy(cp, registry)
	return cp
}

// ResetRegistry clears all registered seeds. Intended for testing.
func ResetRegistry() {
	mu.Lock()
	defer mu.Unlock()

	registry = nil
	nameSet = make(map[string]bool)
}
