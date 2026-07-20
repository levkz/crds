package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKeymapConfig(t *testing.T) {
	data := []byte(`
global:
  quit:
    keys: ["ctrl+q"]
  help:
    help: "help"
list:
  up:
    keys: ["w"]
`)

	path := filepath.Join(t.TempDir(), "keymaps.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadKeymapConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Global == nil {
		t.Fatal("expected Global section")
	}
	if cfg.Global.Quit == nil || len(cfg.Global.Quit.Keys) != 1 || cfg.Global.Quit.Keys[0] != "ctrl+q" {
		t.Errorf("Global.Quit.Keys = %v, want [ctrl+q]", cfg.Global.Quit.Keys)
	}
	if cfg.Global.Help == nil || cfg.Global.Help.Help == nil || *cfg.Global.Help.Help != "help" {
		t.Errorf("Global.Help.Help = %v, want help", *cfg.Global.Help.Help)
	}
	if cfg.List == nil || cfg.List.Up == nil || len(cfg.List.Up.Keys) != 1 || cfg.List.Up.Keys[0] != "w" {
		t.Errorf("List.Up.Keys = %v, want [w]", cfg.List.Up.Keys)
	}
}

func TestLoadKeymapConfigMissingFile(t *testing.T) {
	cfg, err := LoadKeymapConfig("/nonexistent/keymaps.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Error("expected nil for missing file")
	}
}
