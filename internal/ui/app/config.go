package app

// Config holds UI-level preferences and settings.
// Populated from defaults and optionally from a config file at startup.
type Config struct {
	KeyHelp string
	KeyQuit string

	AnimationEnabled bool
	DefaultQuizLimit int
	ThemePath        string
}

func DefaultConfig() Config {
	return Config{
		KeyHelp:           "?",
		KeyQuit:           "ctrl+c",
		AnimationEnabled:  false,
		DefaultQuizLimit:  20,
	}
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
