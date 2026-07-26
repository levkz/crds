package storage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"crds/internal/storage/db"
)

func newFileStore(t *testing.T, dir string) *Store {
	t.Helper()
	store, err := NewStore(filepath.Join(dir, "crds.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func TestCreateReserve(t *testing.T) {
	sharedDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	ctx := context.Background()

	// Write state.yaml
	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("selected_decks:\n  - test_deck\ntheme: dark\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	// Write a deck YAML into decks/ and sync
	decksDir := filepath.Join(sharedDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		t.Fatalf("mkdir decks: %v", err)
	}
	writeTestDeck(t, decksDir)
	if err := store.SyncDecks(decksDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}

	// Record some progress and reviews
	due := timeOnly("2026-07-26 12:00:00")
	if err := store.queries.UpsertProgress(ctx, db.UpsertProgressParams{
		DeckID:    "test_deck",
		EntryID:   "entry_1",
		Reverse:   0,
		Ease:      2.5,
		Interval:  1,
		Due:       &due,
		Correct:   5,
		Incorrect: 1,
	}); err != nil {
		t.Fatalf("UpsertProgress: %v", err)
	}
	if err := store.RecordAnswer("test_deck", "entry_1", 3, false); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}

	if err := store.CreateReserve(sharedDir); err != nil {
		t.Fatalf("CreateReserve: %v", err)
	}

	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, err := os.ReadDir(reserveDir)
	if err != nil {
		t.Fatalf("read reserve dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 reserve file, got %d", len(entries))
	}

	name := entries[0].Name()
	if !strings.HasPrefix(name, "crds-rsv-") || !strings.HasSuffix(name, ".tar.gz") {
		t.Errorf("unexpected filename: %s", name)
	}

	// Verify archive contents
	reservePath := filepath.Join(reserveDir, name)
	f, err := os.Open(reservePath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	found := map[string]bool{}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar next: %v", err)
		}
		found[hdr.Name] = true
	}

	for _, name := range []string{"state.yaml", "crds.db", "decks/test_deck.yaml"} {
		if !found[name] {
			t.Errorf("missing from archive: %s", name)
		}
	}
}

func TestCreateReserve_Increment(t *testing.T) {
	sharedDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	if err := store.CreateReserve(sharedDir); err != nil {
		t.Fatalf("first CreateReserve: %v", err)
	}

	time.Sleep(2 * time.Second)
	if err := store.CreateReserve(sharedDir); err != nil {
		t.Fatalf("second CreateReserve: %v", err)
	}

	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, err := os.ReadDir(reserveDir)
	if err != nil {
		t.Fatalf("read reserve dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 files, got %d", len(entries))
	}

	has001 := false
	has002 := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "-001-") {
			has001 = true
		}
		if strings.Contains(e.Name(), "-002-") {
			has002 = true
		}
	}
	if !has001 {
		t.Error("expected -001- in first backup")
	}
	if !has002 {
		t.Error("expected -002- in second backup")
	}
}

func TestCreateReserve_MissingFiles(t *testing.T) {
	sharedDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	// No state.yaml, no decks, but a DB exists from newFileStore.
	// Should not error — missing files are skipped silently.
	if err := store.CreateReserve(sharedDir); err != nil {
		t.Fatalf("CreateReserve: %v", err)
	}

	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, _ := os.ReadDir(reserveDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
}

func TestCreateReserve_ExtractRoundTrip(t *testing.T) {
	sharedDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("theme: light\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	decksDir := filepath.Join(sharedDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		t.Fatalf("mkdir decks: %v", err)
	}
	writeTestDeck(t, decksDir)
	if err := store.SyncDecks(decksDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}

	if err := store.CreateReserve(sharedDir); err != nil {
		t.Fatalf("CreateReserve: %v", err)
	}

	// Extract into a fresh directory
	extractDir := t.TempDir()
	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, _ := os.ReadDir(reserveDir)
	if len(entries) == 0 {
		t.Fatal("no reserve file")
	}

	reservePath := filepath.Join(reserveDir, entries[0].Name())
	f, err := os.Open(reservePath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar: %v", err)
		}
		target := filepath.Join(extractDir, filepath.FromSlash(hdr.Name))
		if hdr.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		out, err := os.Create(target)
		if err != nil {
			t.Fatalf("create %s: %v", target, err)
		}
		_, err = io.Copy(out, tr)
		out.Close()
		if err != nil {
			t.Fatalf("write %s: %v", target, err)
		}
	}

	for _, name := range []string{"state.yaml", "crds.db", "decks/test_deck.yaml"} {
		full := filepath.Join(extractDir, filepath.FromSlash(name))
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("extracted file missing: %s", name)
		}
	}

	data, _ := os.ReadFile(filepath.Join(extractDir, "state.yaml"))
	if !strings.Contains(string(data), "theme: light") {
		t.Errorf("state.yaml content wrong: %s", data)
	}

	deckData, _ := os.ReadFile(filepath.Join(extractDir, "decks", "test_deck.yaml"))
	if !strings.Contains(string(deckData), "bonjour") {
		t.Errorf("deck content wrong: %s", deckData)
	}
}
