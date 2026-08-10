package ai

import (
	"testing"
)

func TestResolve_DefaultProvider(t *testing.T) {
	t.Setenv("CRDS_AI_PROVIDER", "")
	t.Setenv("CRDS_AI_BASE_URL", "")
	t.Setenv("CRDS_AI_MODEL", "")
	t.Setenv("CRDS_AI_API_KEY", "")

	cfg, err := Resolve(Config{})
	if err != nil {
		t.Fatalf("Resolve zero: %v", err)
	}
	if cfg.Provider != "pollinations" {
		t.Errorf("Provider = %q, want pollinations", cfg.Provider)
	}
	if cfg.BaseURL != Presets["pollinations"].BaseURL {
		t.Errorf("BaseURL = %q, want preset", cfg.BaseURL)
	}
	if cfg.Model != Presets["pollinations"].DefaultModel {
		t.Errorf("Model = %q, want preset default", cfg.Model)
	}
}

func TestResolve_FromConfigFile(t *testing.T) {
	t.Setenv("CRDS_AI_PROVIDER", "")
	t.Setenv("CRDS_AI_BASE_URL", "")
	t.Setenv("CRDS_AI_MODEL", "")
	t.Setenv("CRDS_AI_API_KEY", "sk-test")

	cfg, err := Resolve(Config{Provider: "groq", Model: "my-model"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Provider != "groq" {
		t.Errorf("Provider = %q, want groq", cfg.Provider)
	}
	if cfg.BaseURL != Presets["groq"].BaseURL {
		t.Errorf("BaseURL = %q, want groq preset", cfg.BaseURL)
	}
	if cfg.Model != "my-model" {
		t.Errorf("Model = %q, want config-file override", cfg.Model)
	}
}

func TestResolve_EnvOverrides(t *testing.T) {
	t.Setenv("CRDS_AI_PROVIDER", "ollama")
	t.Setenv("CRDS_AI_BASE_URL", "http://10.0.0.1:1234/v1")
	t.Setenv("CRDS_AI_MODEL", "qwen3:8b")
	t.Setenv("CRDS_AI_API_KEY", "")

	cfg, err := Resolve(Config{Provider: "groq"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Provider != "ollama" {
		t.Errorf("Provider = %q, want ollama from env", cfg.Provider)
	}
	if cfg.BaseURL != "http://10.0.0.1:1234/v1" {
		t.Errorf("BaseURL = %q, want env override", cfg.BaseURL)
	}
	if cfg.Model != "qwen3:8b" {
		t.Errorf("Model = %q, want env override", cfg.Model)
	}
}

func TestResolve_KeyedProviderRequiresKey(t *testing.T) {
	t.Setenv("CRDS_AI_PROVIDER", "")
	t.Setenv("CRDS_AI_API_KEY", "")

	if _, err := Resolve(Config{Provider: "openai"}); err == nil {
		t.Fatal("expected error for keyed provider without key")
	}

	t.Setenv("CRDS_AI_API_KEY", "sk-test")
	if _, err := Resolve(Config{Provider: "openai"}); err != nil {
		t.Fatalf("expected success with key, got %v", err)
	}
}

func TestResolve_KeylessProvidersDoNotNeedKey(t *testing.T) {
	t.Setenv("CRDS_AI_API_KEY", "")

	for _, name := range []string{"pollinations", "ollama"} {
		t.Setenv("CRDS_AI_PROVIDER", name)
		if _, err := Resolve(Config{}); err != nil {
			t.Errorf("Resolve(%q): %v", name, err)
		}
	}
}

func TestResolve_UnknownProvider(t *testing.T) {
	t.Setenv("CRDS_AI_PROVIDER", "")

	if _, err := Resolve(Config{Provider: "not-a-provider"}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestResolve_EnvUnknownProvider(t *testing.T) {
	t.Setenv("CRDS_AI_PROVIDER", "bogus")

	if _, err := Resolve(Config{}); err == nil {
		t.Fatal("expected error for unknown env provider")
	}
}

func TestPresetsRegistered(t *testing.T) {
	for _, name := range ProviderNames {
		if _, ok := Presets[name]; !ok {
			t.Errorf("provider name %q missing from Presets map", name)
		}
	}
}

func TestConfig_EnvKeyPropagates(t *testing.T) {
	// The environment key is applied regardless of provider; keyless providers
	// simply don't *require* one.
	t.Setenv("CRDS_AI_API_KEY", "secret-value")

	cfg, err := Resolve(Config{Provider: "ollama"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.APIKey != "secret-value" {
		t.Errorf("APIKey = %q, want env value", cfg.APIKey)
	}
}