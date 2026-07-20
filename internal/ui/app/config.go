package app

import cfg "crds/internal/config"

// Config holds UI-level preferences and settings.
// Populated from defaults and optionally from a config file at startup.
type Config struct {
	AnimationEnabled bool
	DefaultQuizLimit int
	ThemePath        string
}

func DefaultConfig() Config {
	return Config{
		AnimationEnabled: false,
		DefaultQuizLimit: 20,
	}
}

// ApplyYAML overrides config fields from a parsed ~/.config/crds/config.yaml.
func (c Config) ApplyYAML(y *cfg.ConfigYAML) Config {
	if y == nil {
		return c
	}
	if y.AnimationEnabled != nil {
		c.AnimationEnabled = *y.AnimationEnabled
	}
	if y.DefaultQuizLimit != nil {
		c.DefaultQuizLimit = *y.DefaultQuizLimit
	}
	return c
}

// WithTheme returns a copy of the config with the given theme path set.
func (c Config) WithTheme(path string) Config {
	c.ThemePath = path
	return c
}

// ConfigUpdatedMsg is emitted when the runtime config has changed.
type ConfigUpdatedMsg struct {
	Config Config
}
