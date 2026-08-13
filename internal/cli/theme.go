package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"crds/internal/app"
	"crds/internal/config"
	"crds/internal/editor"
	"crds/internal/ui/theme"
)

type ThemeCmd struct {
	Add    ThemeAddCmd    `cmd:"" help:"Create a new theme."`
	Delete ThemeDeleteCmd `cmd:"" help:"Delete a theme."`
	Edit   ThemeEditCmd   `cmd:"" help:"Edit a theme by opening its YAML file."`
	List   ThemeListCmd   `cmd:"" help:"List all user themes."`
}

// themesDir is a seam for tests; the real implementation reads the config
// directory.
var themesDir = config.ThemesDir

// themePath resolves a theme name to its config-dir YAML path, rejecting
// names that would escape the directory.
func themePath(name string) (string, error) {
	if name == "" || strings.ContainsAny(name, `/\`) ||
		strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
		return "", fmt.Errorf("invalid theme name %q", name)
	}
	dir, err := themesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".yaml"), nil
}

type ThemeAddCmd struct {
	Name   string `arg:"" required:"" help:"Name of the new theme."`
	Preset string `short:"p" completion-predictor:"preset" help:"Built-in preset to start from (default, dark, light, tokyonight, mocha)."`
}

func (c *ThemeAddCmd) Run(a *app.App) error {
	path, err := themePath(c.Name)
	if err != nil {
		return fmt.Errorf("theme add: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("theme add: theme %q already exists at %s", c.Name, path)
	}

	content := themeTemplate()
	if c.Preset != "" {
		data, ok := theme.BuiltinYAML(c.Preset)
		if !ok {
			return fmt.Errorf("theme add: unknown preset %q (available: %s)", c.Preset, strings.Join(theme.BuiltinNames(), ", "))
		}
		content = string(data)
	}

	if _, err := theme.ParseTheme([]byte(content)); err != nil {
		return fmt.Errorf("theme add: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("theme add: write %s: %w", path, err)
	}

	if c.Preset != "" {
		fmt.Printf("Created theme %q from preset %q.\n", c.Name, c.Preset)
	} else {
		fmt.Printf("Created theme %q.\n", c.Name)
	}
	return nil
}

type ThemeDeleteCmd struct {
	Name  string `arg:"" required:"" help:"Theme to delete." completion-predictor:"theme"`
	Force bool   `short:"f" help:"Skip confirmation."`
}

func (c *ThemeDeleteCmd) Run(a *app.App) error {
	path, err := themePath(c.Name)
	if err != nil {
		return fmt.Errorf("theme delete: %w", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("theme delete: theme %q not found", c.Name)
	}

	if !c.Force {
		fmt.Printf("Delete theme %q? [y/N] ", c.Name)
		var answer string
		if _, err := fmt.Scan(&answer); err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("theme delete: %w", err)
	}
	fmt.Printf("Deleted theme %q.\n", c.Name)
	return nil
}

type ThemeEditCmd struct {
	Name string `arg:"" required:"" help:"Theme to edit." completion-predictor:"theme"`
}

func (c *ThemeEditCmd) Run(a *app.App) error {
	path, err := themePath(c.Name)
	if err != nil {
		return fmt.Errorf("theme edit: %w", err)
	}

	seeded, err := seedThemeIfMissing(path, c.Name)
	if err != nil {
		return err
	}
	if seeded {
		fmt.Printf("Theme %q is built-in; copied it to %s for editing.\n", c.Name, path)
	}

	origRaw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("theme edit: read %s: %w", path, err)
	}

	var editedRaw []byte
	for {
		input := origRaw
		if editedRaw != nil {
			input = editedRaw
		}
		result, err := editor.Edit(string(input))
		if err != nil {
			return err
		}
		editedRaw = []byte(result)

		if _, err := theme.ParseTheme(editedRaw); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			switch promptChoice("Discard changes, Continue editing, Save anyway", "d", "c", "s") {
			case "d":
				fmt.Println("Changes discarded.")
				return nil
			case "c":
				continue
			case "s":
				if err := os.WriteFile(path, editedRaw, 0644); err != nil {
					return fmt.Errorf("theme edit: write %s: %w", path, err)
				}
				fmt.Println("Saved (theme may fail to load at startup).")
				return nil
			}
		}

		if err := os.WriteFile(path, editedRaw, 0644); err != nil {
			return fmt.Errorf("theme edit: write %s: %w", path, err)
		}
		fmt.Printf("Theme %q updated.\n", c.Name)
		return nil
	}
}

type ThemeListCmd struct{}

func (c *ThemeListCmd) Run(a *app.App) error {
	dir, err := themesDir()
	if err != nil {
		return fmt.Errorf("theme list: %w", err)
	}
	files, err := config.DiscoverThemeFilesIn(dir)
	if err != nil {
		return fmt.Errorf("theme list: %w", err)
	}
	names := make([]string, 0, len(files))
	for _, tf := range files {
		names = append(names, tf.Name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Println("No user themes found.")
		return nil
	}
	for _, n := range names {
		fmt.Println(n)
	}
	return nil
}

// seedThemeIfMissing writes a copy of the built-in theme to path when the
// file does not exist yet, returning whether it did so. When the name is not
// a built-in either, it returns an error pointing at crds theme add.
func seedThemeIfMissing(path, name string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	data, ok := theme.BuiltinYAML(name)
	if !ok {
		return false, fmt.Errorf("theme edit: theme %q not found (create it with crds theme add)", name)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return false, fmt.Errorf("theme edit: write %s: %w", path, err)
	}
	return true, nil
}

// themeTemplate returns a blank custom theme with every field commented out.
// Unset fields inherit the default theme's values.
func themeTemplate() string {
	return `# CRDS custom theme. Uncomment a line to override the default theme.
# A color value is a palette key name ("blue"), an ANSI 256 number ("39"),
# or a hex value ("#00afff").
palette:
  # blue: "39"
  # green: "42"
  # orange: "214"
  # red: "196"
  # gray: "248"
  # white: "255"
  # background: "0"
  # selection: "27"
  # border: "59"
  # link: "33"
  # surface: "235"
  # magenta: "177"
  # purple: "140"
  # cyan: "117"
  # yellow: "220"
  # primary: "blue"
  # secondary: "cyan"
  # accent: "orange"
# icons: nerdfont
# typography:
#   title:
#     color: primary
#     bold: true
`
}

// promptChoice reads a single-character choice from stdin, looping until a
// valid option is entered. On EOF it returns the first option.
func promptChoice(message string, options ...string) string {
	valid := make(map[string]bool, len(options))
	for _, o := range options {
		valid[o] = true
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", message, strings.Join(options, "/"))
		if !scanner.Scan() {
			return options[0]
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if valid[answer] {
			return answer
		}
		fmt.Fprintf(os.Stderr, "Invalid choice. Valid options: %s\n", strings.Join(options, ", "))
	}
}
