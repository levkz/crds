package app

import (
	cfg "crds/internal/config"
	"crds/internal/fuzzy"
	"crds/internal/mapping"
	"crds/internal/ui"
)

// Config holds UI-level preferences and settings.
// Populated from defaults and optionally from a config file at startup.
type Config struct {
	AnimationEnabled bool
	DefaultQuizLimit int
	ThemePath        string
	QuizMode         ui.QuizMode
	// MatchingMode controls accent handling when grading typed answers.
	MatchingMode fuzzy.Mode
	// Mappings holds per-language input mappings (built-in + user files).
	Mappings *mapping.Store
}

func DefaultConfig() Config {
	return Config{
		AnimationEnabled: false,
		DefaultQuizLimit: 20,
		QuizMode:         ui.QuizModeNormal,
		MatchingMode:     fuzzy.Approximate,
		Mappings:         mapping.NewStore(),
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
	if y.QuizMode != "" {
		c.QuizMode = ui.ParseQuizMode(y.QuizMode)
	}
	if y.MatchingMode != "" {
		c.MatchingMode = fuzzy.ParseMode(y.MatchingMode)
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
