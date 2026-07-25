package storage

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "go.yaml.in/yaml/v3"
)

// State holds ephemeral application state persisted between sessions.
type State struct {
	SelectedDecks []string `yaml:"selected_decks"`
	Theme         string   `yaml:"theme,omitempty"`
}

// StateStore manages loading and saving State to a YAML file.
type StateStore struct {
	path string
}

// NewStateStore creates a StateStore rooted at dir (e.g. ~/.local/share/crds/).
// It creates the directory if it doesn't exist.
func NewStateStore(dir string) *StateStore {
	return &StateStore{path: filepath.Join(dir, "state.yaml")}
}

// Load reads state from the YAML file. If the file doesn't exist, returns an empty State with no error.
func (s *StateStore) Load() (*State, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("reading state: %w", err)
	}
	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}
	return &state, nil
}

// Save writes state to the YAML file.
func (s *StateStore) Save(state *State) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}
	return nil
}
