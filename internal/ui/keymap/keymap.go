package keymap

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Binding struct {
	Keys []string
	Help string
}

func (b Binding) Match(msg tea.KeyMsg) bool {
	s := msg.String()
	for _, k := range b.Keys {
		if s == k {
			return true
		}
	}
	return false
}

type BindingList []Binding

func (bl BindingList) Help() string {
	var labels []string
	for _, b := range bl {
		if b.Help == "" {
			continue
		}
		labels = append(labels, b.Help)
	}
	return strings.Join(labels, " · ")
}

type Global struct {
	Quit Binding
	Help Binding
	Back Binding
}

func (km Global) Footer() string {
	return BindingList{km.Help, km.Back}.Help()
}

type List struct {
	Up     Binding
	Down   Binding
	Select Binding
}

func (km List) Footer() string {
	return BindingList{km.Up, km.Down, km.Select}.Help()
}

type Quiz struct {
	Reveal  Binding
	Again   Binding
	Hard    Binding
	Good    Binding
	Easy    Binding
}

func (km Quiz) Unrevealed() string {
	return BindingList{km.Reveal}.Help()
}

func (km Quiz) Revealed() string {
	return BindingList{km.Again, km.Hard, km.Good, km.Easy}.Help()
}

type Search struct {
	FocusToggle Binding
	Open        Binding
	DeleteChar  Binding
	List
}

func (km Search) Footer() string {
	return BindingList{km.FocusToggle, km.List.Up, km.List.Down, km.Open}.Help()
}

// NamedBinding pairs a Binding with its group and action name for display.
type NamedBinding struct {
	Group   string
	Action  string
	Binding Binding
}

// Registry holds all application keymaps as a single unit.
type Registry struct {
	Global Global
	List   List
	Quiz   Quiz
	Search Search
}

// Bindings returns every registered binding with its group/action label.
func (r Registry) Bindings() []NamedBinding {
	return []NamedBinding{
		{"Global", "Quit", r.Global.Quit},
		{"Global", "Help", r.Global.Help},
		{"Global", "Back", r.Global.Back},
		{"List", "Up", r.List.Up},
		{"List", "Down", r.List.Down},
		{"List", "Select", r.List.Select},
		{"Quiz", "Reveal", r.Quiz.Reveal},
		{"Quiz", "Again", r.Quiz.Again},
		{"Quiz", "Hard", r.Quiz.Hard},
		{"Quiz", "Good", r.Quiz.Good},
		{"Quiz", "Easy", r.Quiz.Easy},
		{"Search", "FocusToggle", r.Search.FocusToggle},
		{"Search", "Open", r.Search.Open},
		{"Search", "DeleteChar", r.Search.DeleteChar},
	}
}

// FindBinding returns the first binding whose Keys include the given key string.
func (r Registry) FindBinding(key string) *NamedBinding {
	for _, nb := range r.Bindings() {
		for _, k := range nb.Binding.Keys {
			if k == key {
				return &nb
			}
		}
	}
	return nil
}

var DefaultGlobal = Global{
	Quit: Binding{Keys: []string{"ctrl+c"}, Help: "ctrl+c quit"},
	Help: Binding{Keys: []string{"?"}, Help: "? help"},
	Back: Binding{Keys: []string{"esc"}, Help: "esc back"},
}

var DefaultList = List{
	Up:     Binding{Keys: []string{"up", "k"}, Help: "↑ navigate"},
	Down:   Binding{Keys: []string{"down", "j"}, Help: "↓ navigate"},
	Select: Binding{Keys: []string{"enter"}, Help: "enter select"},
}

var DefaultQuiz = Quiz{
	Reveal: Binding{Keys: []string{"enter"}, Help: "enter reveal"},
	Again:  Binding{Keys: []string{"1"}, Help: "1 again"},
	Hard:   Binding{Keys: []string{"2"}, Help: "2 hard"},
	Good:   Binding{Keys: []string{"3"}, Help: "3 good"},
	Easy:   Binding{Keys: []string{"4"}, Help: "4 easy"},
}

var DefaultSearch = Search{
	FocusToggle: Binding{Keys: []string{"tab"}, Help: "tab focus"},
	Open:        Binding{Keys: []string{"enter"}, Help: "enter open"},
	DeleteChar:  Binding{Keys: []string{"backspace"}, Help: ""},
	List:        DefaultList,
}

// DefaultRegistry is the default keybinding configuration for the application.
var DefaultRegistry = Registry{
	Global: DefaultGlobal,
	List:   DefaultList,
	Quiz:   DefaultQuiz,
	Search: DefaultSearch,
}

// BindingOverride specifies user-defined keys and/or help text for a Binding.
// Only non-nil fields are applied when overriding.
type BindingOverride struct {
	Keys []string `yaml:"keys"`
	Help *string  `yaml:"help,omitempty"`
}

// KeymapConfig holds optional overrides for every keymap group.
// Each field is a pointer — nil means "keep defaults for this group".
type KeymapConfig struct {
	Global *struct {
		Quit *BindingOverride `yaml:"quit,omitempty"`
		Help *BindingOverride `yaml:"help,omitempty"`
		Back *BindingOverride `yaml:"back,omitempty"`
	} `yaml:"global,omitempty"`
	List *struct {
		Up     *BindingOverride `yaml:"up,omitempty"`
		Down   *BindingOverride `yaml:"down,omitempty"`
		Select *BindingOverride `yaml:"select,omitempty"`
	} `yaml:"list,omitempty"`
	Quiz *struct {
		Reveal *BindingOverride `yaml:"reveal,omitempty"`
		Again  *BindingOverride `yaml:"again,omitempty"`
		Hard   *BindingOverride `yaml:"hard,omitempty"`
		Good   *BindingOverride `yaml:"good,omitempty"`
		Easy   *BindingOverride `yaml:"easy,omitempty"`
	} `yaml:"quiz,omitempty"`
	Search *struct {
		FocusToggle *BindingOverride `yaml:"focus_toggle,omitempty"`
		Open        *BindingOverride `yaml:"open,omitempty"`
		DeleteChar  *BindingOverride `yaml:"delete_char,omitempty"`
	} `yaml:"search,omitempty"`
}

func applyOverride(b Binding, o BindingOverride) Binding {
	if o.Keys != nil {
		b.Keys = o.Keys
	}
	if o.Help != nil {
		b.Help = *o.Help
	}
	return b
}

// ApplyDefaultOverrides updates the package-level DefaultGlobal, DefaultList,
// DefaultQuiz, DefaultSearch, and DefaultRegistry with user-specified overrides.
// Call once during application initialization, before any screens render.
func ApplyDefaultOverrides(cfg KeymapConfig) {
	if cfg.Global != nil {
		if cfg.Global.Quit != nil {
			DefaultGlobal.Quit = applyOverride(DefaultGlobal.Quit, *cfg.Global.Quit)
		}
		if cfg.Global.Help != nil {
			DefaultGlobal.Help = applyOverride(DefaultGlobal.Help, *cfg.Global.Help)
		}
		if cfg.Global.Back != nil {
			DefaultGlobal.Back = applyOverride(DefaultGlobal.Back, *cfg.Global.Back)
		}
	}
	if cfg.List != nil {
		if cfg.List.Up != nil {
			DefaultList.Up = applyOverride(DefaultList.Up, *cfg.List.Up)
		}
		if cfg.List.Down != nil {
			DefaultList.Down = applyOverride(DefaultList.Down, *cfg.List.Down)
		}
		if cfg.List.Select != nil {
			DefaultList.Select = applyOverride(DefaultList.Select, *cfg.List.Select)
		}
	}
	if cfg.Quiz != nil {
		if cfg.Quiz.Reveal != nil {
			DefaultQuiz.Reveal = applyOverride(DefaultQuiz.Reveal, *cfg.Quiz.Reveal)
		}
		if cfg.Quiz.Again != nil {
			DefaultQuiz.Again = applyOverride(DefaultQuiz.Again, *cfg.Quiz.Again)
		}
		if cfg.Quiz.Hard != nil {
			DefaultQuiz.Hard = applyOverride(DefaultQuiz.Hard, *cfg.Quiz.Hard)
		}
		if cfg.Quiz.Good != nil {
			DefaultQuiz.Good = applyOverride(DefaultQuiz.Good, *cfg.Quiz.Good)
		}
		if cfg.Quiz.Easy != nil {
			DefaultQuiz.Easy = applyOverride(DefaultQuiz.Easy, *cfg.Quiz.Easy)
		}
	}
	if cfg.Search != nil {
		if cfg.Search.FocusToggle != nil {
			DefaultSearch.FocusToggle = applyOverride(DefaultSearch.FocusToggle, *cfg.Search.FocusToggle)
		}
		if cfg.Search.Open != nil {
			DefaultSearch.Open = applyOverride(DefaultSearch.Open, *cfg.Search.Open)
		}
		if cfg.Search.DeleteChar != nil {
			DefaultSearch.DeleteChar = applyOverride(DefaultSearch.DeleteChar, *cfg.Search.DeleteChar)
		}
	}
	// Keep DefaultRegistry in sync with the per-group defaults.
	DefaultRegistry = Registry{
		Global: DefaultGlobal,
		List:   DefaultList,
		Quiz:   DefaultQuiz,
		Search: DefaultSearch,
	}
}
