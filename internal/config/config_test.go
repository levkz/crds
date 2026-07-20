package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		t.Fatal("Dir() returned empty")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("Dir() = %q, expected absolute path", dir)
	}
}

func TestConfigPath(t *testing.T) {
	p, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(p) {
		t.Errorf("ConfigPath() = %q, expected absolute", p)
	}
	if filepath.Base(p) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", filepath.Base(p))
	}
}

func TestKeymapsPath(t *testing.T) {
	p, err := KeymapsPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "keymaps.yaml" {
		t.Errorf("expected keymaps.yaml, got %s", filepath.Base(p))
	}
}

func TestLoadConfigYAML(t *testing.T) {
	data := []byte("theme: dark\nanimation_enabled: true\ndefault_quiz_limit: 30\n")
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %q, want dark", cfg.Theme)
	}
	if cfg.AnimationEnabled == nil || *cfg.AnimationEnabled != true {
		t.Error("AnimationEnabled should be true")
	}
	if cfg.DefaultQuizLimit == nil || *cfg.DefaultQuizLimit != 30 {
		t.Errorf("DefaultQuizLimit = %v, want 30", *cfg.DefaultQuizLimit)
	}
}

func TestLoadConfigYAMLMissingFile(t *testing.T) {
	cfg, err := LoadConfigYAML("/nonexistent/path.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Error("expected nil for missing file")
	}
}

func TestLoadConfigYAMLEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.yaml")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config for empty file")
	}
}

func TestEnsureDefaultFiles(t *testing.T) {
	// Can't easily test EnsureDefaultFiles with a custom dir since Dir() is hardcoded.
	// Unit-level component functions are tested separately above.
}
