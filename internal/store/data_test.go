package store

import "testing"

// minRecordsPerType is the floor every bundled dataset must meet, guarding
// against a data file silently shrinking below the intended sample size.
const minRecordsPerType = 100

// TestBundledDatasetsMeetMinimumSize loads the repository's data directory and
// asserts every discovered entity type carries at least minRecordsPerType
// records. It doubles as a smoke test that every data/*.json file parses and
// adapts cleanly through LoadDir.
func TestBundledDatasetsMeetMinimumSize(t *testing.T) {
	data, err := LoadDir("../../data")
	if err != nil {
		t.Fatalf("loading bundled data: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("no datasets found under ../../data")
	}

	for name, items := range data {
		if len(items) < minRecordsPerType {
			t.Errorf("dataset %q has %d records, want at least %d", name, len(items), minRecordsPerType)
		}
	}
}
