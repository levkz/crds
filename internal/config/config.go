package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "go.yaml.in/yaml/v3"
)

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".config", "crds"), nil
}

func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}
	return dir, nil
}

func ConfigPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func KeymapsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "keymaps.yaml"), nil
}

// themesDir returns the path to ~/.config/crds/themes/, creating it if needed.
// Override in tests by assigning a new function.
var themesDir = func() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "themes")
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", fmt.Errorf("creating %s: %w", p, err)
	}
	return p, nil
}

func ThemesDir() (string, error) {
	return themesDir()
}

type ConfigYAML struct {
	Theme            string `yaml:"theme,omitempty"`
	AnimationEnabled *bool  `yaml:"animation_enabled,omitempty"`
	DefaultQuizLimit *int   `yaml:"default_quiz_limit,omitempty"`
	QuizMode         string `yaml:"quiz_mode,omitempty"`
}

func LoadConfigYAML(path string) (*ConfigYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg ConfigYAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

func writeDefaultConfig(path string) error {
	data := []byte("# crds application config\n# theme: dark\n# animation_enabled: false\n# default_quiz_limit: 20\n")
	return os.WriteFile(path, data, 0644)
}

func writeDefaultKeymaps(path string) error {
	data := []byte("# crds keybinding overrides\n# global:\n#   quit:\n#     keys: [\"ctrl+q\"]\n")
	return os.WriteFile(path, data, 0644)
}

func EnsureDefaultFiles() error {
	dir, err := EnsureDir()
	if err != nil {
		return err
	}

	cfgPath := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := writeDefaultConfig(cfgPath); err != nil {
			return fmt.Errorf("writing %s: %w", cfgPath, err)
		}
	}

	keyPath := filepath.Join(dir, "keymaps.yaml")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if err := writeDefaultKeymaps(keyPath); err != nil {
			return fmt.Errorf("writing %s: %w", keyPath, err)
		}
	}

	if _, err := ThemesDir(); err != nil {
		return err
	}

	return nil
}
