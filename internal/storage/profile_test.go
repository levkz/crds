package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateProfile_RoundTrip(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	// Write state.yaml
	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("theme: dark\nselected_decks:\n  - test_deck\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	// Write a deck
	decksDir := filepath.Join(sharedDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		t.Fatalf("mkdir decks: %v", err)
	}
	writeTestDeck(t, decksDir)
	if err := store.SyncDecks(decksDir); err != nil {
		t.Fatalf("SyncDecks: %v", err)
	}

	// Write config files
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("theme: tokyonight\n"), 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "keymaps.yaml"), []byte("# keybindings\n"), 0644); err != nil {
		t.Fatalf("write keymaps.yaml: %v", err)
	}
	themesDir := filepath.Join(configDir, "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatalf("mkdir themes: %v", err)
	}
	if err := os.WriteFile(filepath.Join(themesDir, "mytheme.yaml"), []byte("palette:\n  blue: '#89b4fa'\n"), 0644); err != nil {
		t.Fatalf("write theme: %v", err)
	}

	outputDir := t.TempDir()

	// Export profile
	path, err := store.CreateProfile(sharedDir, configDir, outputDir, "")
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if !strings.HasPrefix(filepath.Base(path), "crds-profile") || !strings.HasSuffix(path, ".tar.gz") {
		t.Errorf("unexpected profile filename: %s", filepath.Base(path))
	}

	// Remember original deck yaml content for later comparison
	origDeckData, err := os.ReadFile(filepath.Join(decksDir, "test_deck.yaml"))
	if err != nil {
		t.Fatalf("read original deck: %v", err)
	}
	origConfigData, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read original config: %v", err)
	}
	origThemeData, err := os.ReadFile(filepath.Join(themesDir, "mytheme.yaml"))
	if err != nil {
		t.Fatalf("read original theme: %v", err)
	}

	// Destroy originals
	if err := os.Remove(filepath.Join(sharedDir, "state.yaml")); err != nil {
		t.Fatalf("remove state.yaml: %v", err)
	}
	if err := os.Remove(filepath.Join(decksDir, "test_deck.yaml")); err != nil {
		t.Fatalf("remove deck: %v", err)
	}
	if err := os.Remove(filepath.Join(configDir, "config.yaml")); err != nil {
		t.Fatalf("remove config.yaml: %v", err)
	}
	if err := os.Remove(filepath.Join(configDir, "keymaps.yaml")); err != nil {
		t.Fatalf("remove keymaps.yaml: %v", err)
	}
	if err := os.RemoveAll(themesDir); err != nil {
		t.Fatalf("remove themes: %v", err)
	}

	// Create fresh config dir for import (import creates files/dirs as needed)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir configDir: %v", err)
	}

	// Close and re-create store to simulate fresh start
	store.Close()

	// Reopen to have a fresh store for import
	store2 := newFileStore(t, sharedDir)
	defer store2.Close()

	// Import profile
	if err := store2.ImportProfile(sharedDir, configDir, path); err != nil {
		t.Fatalf("ImportProfile: %v", err)
	}

	// Verify state.yaml restored
	if _, err := os.Stat(filepath.Join(sharedDir, "state.yaml")); os.IsNotExist(err) {
		t.Error("state.yaml not restored")
	}

	// Verify deck restored
	restoredDeckData, err := os.ReadFile(filepath.Join(decksDir, "test_deck.yaml"))
	if err != nil {
		t.Fatalf("read restored deck: %v", err)
	}
	if string(restoredDeckData) != string(origDeckData) {
		t.Errorf("restored deck content differs from original")
	}

	// Verify config.yaml restored
	restoredConfigData, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read restored config: %v", err)
	}
	if string(restoredConfigData) != string(origConfigData) {
		t.Errorf("restored config content differs: got %q, want %q", restoredConfigData, origConfigData)
	}

	// Verify theme restored
	restoredThemeData, err := os.ReadFile(filepath.Join(configDir, "themes", "mytheme.yaml"))
	if err != nil {
		t.Fatalf("read restored theme: %v", err)
	}
	if string(restoredThemeData) != string(origThemeData) {
		t.Errorf("restored theme content differs")
	}

	// Verify DB is usable
	if err := store2.RecordAnswer("test_deck", "entry_1", 3, false); err != nil {
		t.Fatalf("RecordAnswer after import: %v", err)
	}
	stats := store2.Stats()
	if stats.ReviewedToday != 1 {
		t.Errorf("expected 1 review after import, got %d", stats.ReviewedToday)
	}
}

func TestCreateProfile_Naming(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	// Write minimal files
	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	outputDir := t.TempDir()

	// First export — should be crds-profile.tar.gz
	p1, err := store.CreateProfile(sharedDir, configDir, outputDir, "")
	if err != nil {
		t.Fatalf("first CreateProfile: %v", err)
	}
	if filepath.Base(p1) != "crds-profile.tar.gz" {
		t.Errorf("expected crds-profile.tar.gz, got %s", filepath.Base(p1))
	}

	// Second export — should be crds-profile-1.tar.gz
	p2, err := store.CreateProfile(sharedDir, configDir, outputDir, "")
	if err != nil {
		t.Fatalf("second CreateProfile: %v", err)
	}
	if filepath.Base(p2) != "crds-profile-1.tar.gz" {
		t.Errorf("expected crds-profile-1.tar.gz, got %s", filepath.Base(p2))
	}

	// Third — should be crds-profile-2.tar.gz
	p3, err := store.CreateProfile(sharedDir, configDir, outputDir, "")
	if err != nil {
		t.Fatalf("third CreateProfile: %v", err)
	}
	if filepath.Base(p3) != "crds-profile-2.tar.gz" {
		t.Errorf("expected crds-profile-2.tar.gz, got %s", filepath.Base(p3))
	}
}

func TestCreateProfile_CustomName(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	outputDir := t.TempDir()

	path, err := store.CreateProfile(sharedDir, configDir, outputDir, "mymigration")
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if filepath.Base(path) != "mymigration.tar.gz" {
		t.Errorf("expected mymigration.tar.gz, got %s", filepath.Base(path))
	}
}

func TestImportProfile_InvalidArchive(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	badPath := filepath.Join(t.TempDir(), "bad.tar.gz")
	if err := os.WriteFile(badPath, []byte("not gzip data"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	err := store.ImportProfile(sharedDir, configDir, badPath)
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestCreateProfile_NoConfigFiles(t *testing.T) {
	sharedDir := t.TempDir()
	configDir := t.TempDir()
	store := newFileStore(t, sharedDir)
	defer store.Close()

	if err := os.WriteFile(filepath.Join(sharedDir, "state.yaml"), []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	outputDir := t.TempDir()

	// No configDir files at all — should not error
	if _, err := store.CreateProfile(sharedDir, configDir, outputDir, ""); err != nil {
		t.Fatalf("CreateProfile without config files: %v", err)
	}
}
