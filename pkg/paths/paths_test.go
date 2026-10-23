package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesAndReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "file.json")

	if err := WriteFileAtomic(path, []byte("one"), 0644); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "one" {
		t.Fatalf("content = %q, want one", string(data))
	}

	if err := WriteFileAtomic(path, []byte("two"), 0644); err != nil {
		t.Fatalf("replace failed: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "two" {
		t.Fatalf("content = %q, want two", string(data))
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}
