package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"crds/internal/model"

	"go.yaml.in/yaml/v3"
)

func pickEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	for _, name := range []string{"nano", "vim", "vi"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return "vi"
}

// Edit opens the user's preferred editor with the given content and returns
// the modified content after the editor exits. Uses $EDITOR, falling back to
// nano, vim, then vi.
func Edit(content string) (string, error) {
	dir, err := os.MkdirTemp("", "crds-edit-*")
	if err != nil {
		return "", fmt.Errorf("editor: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	tmp := filepath.Join(dir, "entry.yaml")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("editor: write: %w", err)
	}

	editor := pickEditor()
	cmd := exec.Command(editor, tmp)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor: %s: %w", editor, err)
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		return "", fmt.Errorf("editor: read: %w", err)
	}

	return string(data), nil
}

// EditEntry marshals an entry to YAML, opens the editor, parses the result,
// and returns the modified entry.
func EditEntry(entry *model.Entry) (*model.Entry, error) {
	data, err := yaml.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("editor: marshal: %w", err)
	}

	modified, err := Edit(string(data))
	if err != nil {
		return nil, err
	}

	var result model.Entry
	if err := yaml.Unmarshal([]byte(modified), &result); err != nil {
		return nil, fmt.Errorf("editor: parse: %w", err)
	}

	return &result, nil
}

// EntryTemplate returns a blank entry as YAML, ready for editing.
func EntryTemplate() string {
	return `# Leave id empty to auto-generate, or set it explicitly.
id: ""
term: ""
translations:
  - text: ""
`
}
