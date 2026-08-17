package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gomock/internal/entity"
)

// LoadDir scans dir for top-level *.json files. Each file's name (without
// its extension) becomes an entity type, and its contents must be a JSON
// array of flat objects that AdaptRecord can normalize. Any failure aborts
// loading with an error naming the offending file and record index.
func LoadDir(dir string) (map[string][]entity.Entity, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading data directory %q: %w", dir, err)
	}

	result := make(map[string][]entity.Entity)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		typeName := strings.TrimSuffix(e.Name(), ".json")
		path := filepath.Join(dir, e.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %q: %w", path, err)
		}

		var records []map[string]any
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("parsing %q: expected a JSON array of objects: %w", path, err)
		}

		entities := make([]entity.Entity, 0, len(records))
		for i, record := range records {
			ent, err := AdaptRecord(record)
			if err != nil {
				return nil, fmt.Errorf("%s: record %d: %w", path, i, err)
			}
			entities = append(entities, ent)
		}

		result[typeName] = entities
	}

	return result, nil
}
