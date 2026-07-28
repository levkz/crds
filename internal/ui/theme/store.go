package theme

import (
	"fmt"
	"sort"
)

type Store struct {
	themes  map[string]Theme
	current string
}

func NewStore() *Store {
	s := &Store{
		themes: make(map[string]Theme),
	}
	s.themes["default"] = Default
	s.themes["dark"] = DarkTheme()
	s.themes["light"] = LightTheme()
	s.themes["tokyonight"] = TokyonightTheme()
	s.themes["mocha"] = MochaTheme()
	s.current = "default"
	return s
}

func (s *Store) Register(name string, t Theme) {
	s.themes[name] = t
}

func (s *Store) RegisterPath(name, path string) error {
	t, err := LoadTheme(path)
	if err != nil {
		return fmt.Errorf("register %q: %w", name, err)
	}
	s.themes[name] = t
	return nil
}

func (s *Store) Names() []string {
	names := make([]string, 0, len(s.themes))
	for name := range s.themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Store) Current() Theme {
	return s.themes[s.current]
}

func (s *Store) CurrentName() string {
	return s.current
}

func (s *Store) Switch(name string) (Theme, error) {
	t, ok := s.themes[name]
	if !ok {
		return Theme{}, fmt.Errorf("theme %q not found", name)
	}
	s.current = name
	return t, nil
}

func (s *Store) Get(name string) (Theme, bool) {
	t, ok := s.themes[name]
	return t, ok
}

func (s *Store) Unregister(name string) {
	delete(s.themes, name)
	if s.current == name {
		s.current = "default"
	}
}

func (s *Store) Has(name string) bool {
	_, ok := s.themes[name]
	return ok
}

func (s *Store) Len() int {
	return len(s.themes)
}

var DefaultStore = NewStore()

func Register(name string, t Theme)        { DefaultStore.Register(name, t) }
func RegisterPath(name, path string) error { return DefaultStore.RegisterPath(name, path) }
func Names() []string                      { return DefaultStore.Names() }
func Current() Theme                       { return DefaultStore.Current() }
func CurrentName() string                  { return DefaultStore.CurrentName() }
func Switch(name string) (Theme, error)    { return DefaultStore.Switch(name) }
