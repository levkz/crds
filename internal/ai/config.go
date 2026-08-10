package ai

import (
	"fmt"
	"os"
)

// DefaultProvider is used when neither config.yaml nor the environment
// sets a provider. Pollinations is chosen for its keyless, no-signup tier.
const DefaultProvider = "pollinations"

// ProviderNames lists the supported presets, in the order shown to users.
var ProviderNames = []string{"pollinations", "ollama", "openai", "gemini", "openrouter", "groq", "nvidia"}

// Preset holds the baked-in defaults for one provider.
type Preset struct {
	BaseURL      string
	DefaultModel string
	RequiresKey  bool
}

// Presets maps provider names to their defaults. Every provider speaks the
// OpenAI chat-completions wire format, so a single client covers all of them.
var Presets = map[string]Preset{
	"pollinations": {BaseURL: "https://text.pollinations.ai/openai", DefaultModel: "openai"},
	"ollama":       {BaseURL: "http://localhost:11434/v1", DefaultModel: "llama3.2"},
	"openai":       {BaseURL: "https://api.openai.com/v1", DefaultModel: "gpt-4o-mini", RequiresKey: true},
	"gemini":       {BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/", DefaultModel: "gemini-3.5-flash-lite", RequiresKey: true},
	"openrouter":   {BaseURL: "https://openrouter.ai/api/v1", DefaultModel: "meta-llama/llama-3.3-70b-instruct:free", RequiresKey: true},
	"groq":         {BaseURL: "https://api.groq.com/openai/v1", DefaultModel: "llama-3.3-70b-versatile", RequiresKey: true},
	"nvidia":       {BaseURL: "https://integrate.api.nvidia.com/v1", DefaultModel: "meta/llama-3.3-70b-instruct", RequiresKey: true},
}

// Config is the effective AI configuration used by the client. A zero value
// means "nothing configured" — Resolve fills in the defaults.
type Config struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

// Resolve returns the effective configuration. Values already set in cfg (from
// config.yaml) are kept; the environment overrides them; otherwise provider
// preset defaults apply. Keyed providers fail without an API key.
func Resolve(cfg Config) (Config, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = DefaultProvider
	}
	if v := os.Getenv("CRDS_AI_PROVIDER"); v != "" {
		provider = v
	}

	preset, ok := Presets[provider]
	if !ok {
		return Config{}, fmt.Errorf("unknown ai provider %q (known: %v)", provider, ProviderNames)
	}

	resolved := Config{Provider: provider}

	resolved.BaseURL = cfg.BaseURL
	if v := os.Getenv("CRDS_AI_BASE_URL"); v != "" {
		resolved.BaseURL = v
	}
	if resolved.BaseURL == "" {
		resolved.BaseURL = preset.BaseURL
	}

	resolved.Model = cfg.Model
	if v := os.Getenv("CRDS_AI_MODEL"); v != "" {
		resolved.Model = v
	}
	if resolved.Model == "" {
		resolved.Model = preset.DefaultModel
	}

	resolved.APIKey = cfg.APIKey
	if v := os.Getenv("CRDS_AI_API_KEY"); v != "" {
		resolved.APIKey = v
	}

	if preset.RequiresKey && resolved.APIKey == "" {
		return Config{}, fmt.Errorf("ai provider %q requires an API key (set ai.api_key in config.yaml or CRDS_AI_API_KEY)", provider)
	}

	return resolved, nil
}

