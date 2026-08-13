package cli

import (
	"os"
	"path/filepath"
	"testing"

	"crds/internal/ui/theme"
)

func withThemesDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "themes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	saved := themesDir
	themesDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { themesDir = saved })
	return dir
}

func TestThemePathValid(t *testing.T) {
	dir := withThemesDir(t)

	cases := []struct {
		name string
		ok   bool
	}{
		{"solarized", true},
		{"my-theme", true},
		{"", false},
		{"a/b", false},
		{"a\\b", false},
		{"foo.yaml", false},
		{"foo.yml", false},
		{"../escape", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, err := themePath(tc.name)
			if tc.ok {
				if err != nil {
					t.Fatalf("themePath(%q): %v", tc.name, err)
				}
				if want := filepath.Join(dir, tc.name+".yaml"); path != want {
					t.Errorf("got %q, want %q", path, want)
				}
			} else if err == nil {
				t.Errorf("themePath(%q) should error", tc.name)
			}
		})
	}
}

func TestThemeAddCmd_Empty(t *testing.T) {
	withThemesDir(t)
	a := newTestApp(t)

	c := &ThemeAddCmd{Name: "solarized"}
	if err := c.Run(a); err != nil {
		t.Fatalf("ThemeAddCmd.Run: %v", err)
	}

	path := filepath.Join(mustThemesDir(t), "solarized.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := theme.ParseTheme(data); err != nil {
		t.Fatalf("created theme does not parse: %v", err)
	}
}

func mustThemesDir(t *testing.T) string {
	dir, err := themesDir()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestThemeAddCmd_FromPreset(t *testing.T) {
	withThemesDir(t)
	a := newTestApp(t)

	c := &ThemeAddCmd{Name: "darkcopy", Preset: "dark"}
	if err := c.Run(a); err != nil {
		t.Fatalf("ThemeAddCmd.Run: %v", err)
	}

	path := filepath.Join(mustThemesDir(t), "darkcopy.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := theme.BuiltinYAML("dark")
	if string(data) != string(want) {
		t.Errorf("preset copy mismatch:\n%s", data)
	}
}

func TestThemeAddCmd_UnknownPreset(t *testing.T) {
	withThemesDir(t)
	a := newTestApp(t)

	c := &ThemeAddCmd{Name: "weird", Preset: "nope"}
	if err := c.Run(a); err == nil {
		t.Fatal("expected error for unknown preset")
	}
}

func TestThemeAddCmd_AlreadyExists(t *testing.T) {
	dir := withThemesDir(t)
	a := newTestApp(t)

	if err := os.WriteFile(filepath.Join(dir, "dark.yaml"), []byte("icons: fallback"), 0644); err != nil {
		t.Fatal(err)
	}
	c := &ThemeAddCmd{Name: "dark"}
	if err := c.Run(a); err == nil {
		t.Fatal("expected error when theme already exists")
	}
}

func TestThemeAddCmd_InvalidName(t *testing.T) {
	withThemesDir(t)
	a := newTestApp(t)

	for _, name := range []string{"", "a/b", "a.yaml"} {
		c := &ThemeAddCmd{Name: name}
		if err := c.Run(a); err == nil {
			t.Errorf("expected error for name %q", name)
		}
	}
}

func TestThemeDeleteCmd_NotFound(t *testing.T) {
	withThemesDir(t)
	a := newTestApp(t)

	c := &ThemeDeleteCmd{Name: "missing", Force: true}
	if err := c.Run(a); err == nil {
		t.Fatal("expected error for missing theme")
	}
}

func TestThemeDeleteCmd_Force(t *testing.T) {
	dir := withThemesDir(t)
	a := newTestApp(t)

	path := filepath.Join(dir, "dark.yaml")
	if err := os.WriteFile(path, []byte("icons: fallback"), 0644); err != nil {
		t.Fatal(err)
	}
	c := &ThemeDeleteCmd{Name: "dark", Force: true}
	if err := c.Run(a); err != nil {
		t.Fatalf("ThemeDeleteCmd.Run: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("theme file was not deleted")
	}
}

func TestThemeEditCmd_SeedsBuiltin(t *testing.T) {
	dir := withThemesDir(t)
	a := newTestApp(t)

	// seedThemeIfMissing only; the full Run opens $EDITOR which is not
	// exercised here.
	path := filepath.Join(dir, "dark.yaml")
	seeded, err := seedThemeIfMissing(path, "dark")
	if err != nil {
		t.Fatalf("seedThemeIfMissing: %v", err)
	}
	if !seeded {
		t.Fatal("expected the built-in theme to be seeded")
	}
	want, _ := theme.BuiltinYAML("dark")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(want) {
		t.Error("seeded content does not match built-in dark theme")
	}

	seeded2, err := seedThemeIfMissing(path, "dark")
	if err != nil {
		t.Fatalf("second seed: %v", err)
	}
	if seeded2 {
		t.Error("second seed should not rewrite an existing file")
	}

	_ = a
}

func TestThemeEditCmd_NotFound(t *testing.T) {
	withThemesDir(t)

	path := filepath.Join(mustThemesDir(t), "nope.yaml")
	if _, err := seedThemeIfMissing(path, "nope"); err == nil {
		t.Fatal("expected error for non-built-in theme name")
	}
}

func TestThemeListCmd(t *testing.T) {
	dir := withThemesDir(t)
	a := newTestApp(t)

	for _, f := range []string{"dark.yaml", "solarized.yaml", "light.yaml", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	c := &ThemeListCmd{}
	if err := c.Run(a); err != nil {
		t.Fatalf("ThemeListCmd.Run: %v", err)
	}
}
