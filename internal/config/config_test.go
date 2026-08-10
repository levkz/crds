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

func TestLoadConfigYAML_AIBlock(t *testing.T) {
	data := []byte("ai:\n  provider: groq\n  model: llama-3.3-70b-versatile\n  api_key: sk-secret\n  base_url: https://example.com/v1\n")
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI == nil {
		t.Fatal("expected ai block to be parsed")
	}
	if cfg.AI.Provider != "groq" {
		t.Errorf("Provider = %q, want groq", cfg.AI.Provider)
	}
	if cfg.AI.Model != "llama-3.3-70b-versatile" {
		t.Errorf("Model = %q", cfg.AI.Model)
	}
	if cfg.AI.APIKey != "sk-secret" {
		t.Errorf("APIKey = %q", cfg.AI.APIKey)
	}
	if cfg.AI.BaseURL != "https://example.com/v1" {
		t.Errorf("BaseURL = %q", cfg.AI.BaseURL)
	}
}

func TestLoadConfigYAML_AIBlockAbsent(t *testing.T) {
	data := []byte("theme: dark\n")
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI != nil {
		t.Errorf("expected nil ai block, got %+v", cfg.AI)
	}
}

func TestLoadConfigYAML_MatchingMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("matching_mode: strict\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MatchingMode != "strict" {
		t.Errorf("MatchingMode = %q, want strict", cfg.MatchingMode)
	}
}

func TestLoadConfigYAML_MatchingModeAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("theme: dark\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfigYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MatchingMode != "" {
		t.Errorf("MatchingMode = %q, want empty default", cfg.MatchingMode)
	}
}

func TestMappingsDir(t *testing.T) {
	p, err := MappingsDir()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(p) {
		t.Errorf("MappingsDir() = %q, expected absolute", p)
	}
	if filepath.Base(p) != "mappings" {
		t.Errorf("expected mappings dir, got %s", filepath.Base(p))
	}
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		t.Errorf("MappingsDir() did not create a directory: %v", err)
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
