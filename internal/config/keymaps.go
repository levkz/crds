package config

import (
	"fmt"
	"os"

	yaml "go.yaml.in/yaml/v3"
	"crds/internal/ui/keymap"
)

func LoadKeymapConfig(path string) (*keymap.KeymapConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg keymap.KeymapConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}
