package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverThemeFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "crds", "themes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	files := []string{"dark.yaml", "light.yaml", "solarized.yaml", "notes.txt", "README.md"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Override themesDir for test
	saved := themesDir
	themesDir = func() (string, error) { return dir, nil }
	defer func() { themesDir = saved }()

	result, err := DiscoverThemeFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 theme files, got %d: %v", len(result), result)
	}
	names := map[string]bool{}
	for _, tf := range result {
		names[tf.Name] = true
	}
	if !names["dark"] {
		t.Error("missing dark theme")
	}
	if !names["light"] {
		t.Error("missing light theme")
	}
	if !names["solarized"] {
		t.Error("missing solarized theme")
	}
}

func TestDiscoverThemeFilesNoDir(t *testing.T) {
	themesDir = func() (string, error) { return "/nonexistent/crds/themes", nil }
	defer func() { themesDir = func() (string, error) { return "", nil } }()

	result, err := DiscoverThemeFiles()
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestDiscoverThemeFilesEmptyDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty_themes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	themesDir = func() (string, error) { return dir, nil }
	defer func() { themesDir = func() (string, error) { return "", nil } }()

	result, err := DiscoverThemeFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 themes, got %d", len(result))
	}
}
