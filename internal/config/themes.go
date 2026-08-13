package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"crds/internal/ui/theme"
)

type ThemeFile struct {
	Name string
	Path string
}

func DiscoverThemeFiles() ([]ThemeFile, error) {
	td, err := themesDir()
	if err != nil {
		return nil, err
	}
	return DiscoverThemeFilesIn(td)
}

// DiscoverThemeFilesIn lists theme YAML files in a specific directory,
// returning each file's stem as the theme name.
func DiscoverThemeFilesIn(td string) ([]ThemeFile, error) {
	entries, err := os.ReadDir(td)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", td, err)
	}

	var files []ThemeFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		files = append(files, ThemeFile{
			Name: base,
			Path: filepath.Join(td, name),
		})
	}
	return files, nil
}

func LoadUserThemes() error {
	files, err := DiscoverThemeFiles()
	if err != nil {
		return err
	}
	for _, tf := range files {
		if err := theme.RegisterPath(tf.Name, tf.Path); err != nil {
			return fmt.Errorf("registering theme %q: %w", tf.Name, err)
		}
	}
	return nil
}
