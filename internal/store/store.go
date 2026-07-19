package store

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"gomock/internal/entity"
)

// TypeInfo describes a discovered entity type and how many records it holds.
type TypeInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Store is a thread-safe, in-memory collection of entities grouped by
// user-defined type name.
type Store struct {
	mu   sync.RWMutex
	data map[string][]entity.Entity
}

// New creates a Store seeded with the given initial data (typically the
// result of LoadDir).
func New(initial map[string][]entity.Entity) *Store {
	data := make(map[string][]entity.Entity, len(initial))
	for k, v := range initial {
		data[k] = v
	}
	return &Store{data: data}
}

// Types returns all known entity types sorted alphabetically by name.
func (s *Store) Types() []TypeInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	types := make([]TypeInfo, 0, len(s.data))
	for name, items := range s.data {
		types = append(types, TypeInfo{Name: name, Count: len(items)})
	}
	sort.Slice(types, func(i, j int) bool { return types[i].Name < types[j].Name })
	return types
}

// List returns all entities for typeName. The second return value reports
// whether the type exists at all.
func (s *Store) List(typeName string) ([]entity.Entity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, ok := s.data[typeName]
	return items, ok
}

// Get looks up a single entity within typeName by numeric ID or, failing
// that, by a case-insensitive exact slug match. The returned booleans are,
// in order: whether a matching record was found, and whether typeName
// exists at all (letting handlers distinguish an unknown type from an
// unknown record within a known type).
func (s *Store) Get(typeName, idOrSlug string) (item entity.Entity, found bool, typeExists bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, typeExists := s.data[typeName]
	if !typeExists {
		return entity.Entity{}, false, false
	}

	if id, err := strconv.Atoi(idOrSlug); err == nil {
		for _, candidate := range items {
			if candidate.ID == id {
				return candidate, true, true
			}
		}
	}

	for _, candidate := range items {
		if strings.EqualFold(candidate.Slug, idOrSlug) {
			return candidate, true, true
		}
	}

	return entity.Entity{}, false, true
}

// Replace overwrites (or creates) the dataset for typeName with items.
// This is in-memory only; it does not persist back to the source JSON
// file, so a server restart reverts to the on-disk baseline.
func (s *Store) Replace(typeName string, items []entity.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[typeName] = items
}
