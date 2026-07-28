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
	Inverse Binding
}

func (km Quiz) Unrevealed() string {
	return BindingList{km.Reveal, km.Inverse}.Help()
}

func (km Quiz) Revealed() string {
	return BindingList{km.Again, km.Hard, km.Good, km.Easy, km.Inverse}.Help()
}

type TypingQuiz struct {
	Submit       Binding
	Reveal       Binding
	Inverse      Binding
	PrevExample  Binding
	NextExample  Binding
}

func (km TypingQuiz) Footer() string {
	return BindingList{km.Submit, km.Reveal, km.Inverse, km.PrevExample, km.NextExample}.Help()
}

func (km TypingQuiz) ExamplesFooter() string {
	return BindingList{km.Inverse, km.PrevExample, km.NextExample}.Help()
}

type Decks struct {
	List
	Toggle    Binding
	ToggleAll Binding
}

func (km Decks) Footer() string {
	return BindingList{km.Up, km.Down, km.Toggle, km.ToggleAll, km.Select}.Help()
}

type Search struct {
	Open       Binding
	DeleteChar Binding
	List
}

func (km Search) Footer() string {
	return BindingList{km.List.Up, km.List.Down, km.Open}.Help()
}

type Home struct {
	Left  Binding
	Right Binding

	FlashCards    Binding
	TypingQuiz    Binding
	Statistics    Binding
	Search        Binding
	Configuration Binding
	DeckSelect    Binding
	Palette       Binding
}

// NamedBinding pairs a Binding with its group and action name for display.
type NamedBinding struct {
	Group   string
	Action  string
	Binding Binding
}

// Registry holds all application keymaps as a single unit.
type Registry struct {
	Global     Global
	List       List
	Home       Home
	Quiz       Quiz
	TypingQuiz TypingQuiz
	Decks      Decks
	Search     Search
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
		{"Home", "Left", r.Home.Left},
		{"Home", "Right", r.Home.Right},
		{"Home", "FlashCards", r.Home.FlashCards},
		{"Home", "TypingQuiz", r.Home.TypingQuiz},
		{"Home", "Statistics", r.Home.Statistics},
		{"Home", "Search", r.Home.Search},
		{"Home", "Configuration", r.Home.Configuration},
		{"Home", "DeckSelect", r.Home.DeckSelect},
		{"Home", "Palette", r.Home.Palette},
		{"Quiz", "Reveal", r.Quiz.Reveal},
		{"Quiz", "Again", r.Quiz.Again},
		{"Quiz", "Hard", r.Quiz.Hard},
		{"Quiz", "Good", r.Quiz.Good},
		{"Quiz", "Easy", r.Quiz.Easy},
		{"Quiz", "Inverse", r.Quiz.Inverse},
		{"TypingQuiz", "Submit", r.TypingQuiz.Submit},
		{"TypingQuiz", "Reveal", r.TypingQuiz.Reveal},
		{"TypingQuiz", "Inverse", r.TypingQuiz.Inverse},
		{"TypingQuiz", "PrevExample", r.TypingQuiz.PrevExample},
		{"TypingQuiz", "NextExample", r.TypingQuiz.NextExample},
		{"Decks", "Toggle", r.Decks.Toggle},
		{"Decks", "ToggleAll", r.Decks.ToggleAll},
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
	Reveal:  Binding{Keys: []string{"enter"}, Help: "enter reveal"},
	Again:   Binding{Keys: []string{"1"}, Help: "1 again"},
	Hard:    Binding{Keys: []string{"2"}, Help: "2 hard"},
	Good:    Binding{Keys: []string{"3"}, Help: "3 good"},
	Easy:    Binding{Keys: []string{"4"}, Help: "4 easy"},
	Inverse: Binding{Keys: []string{"tab"}, Help: "tab inverse"},
}

var DefaultTypingQuiz = TypingQuiz{
	Submit:       Binding{Keys: []string{"enter"},     Help: "enter submit"},
	Reveal:       Binding{Keys: []string{"ctrl+r"},    Help: "ctrl+r reveal"},
	Inverse:      Binding{Keys: []string{"tab"},       Help: "tab inverse"},
	PrevExample:  Binding{Keys: []string{"left", "["}, Help: "[ previous"},
	NextExample:  Binding{Keys: []string{"right", "]"},Help: "] next"},
}

var DefaultDecks = Decks{
	List:      DefaultList,
	Toggle:    Binding{Keys: []string{" "}, Help: "space toggle"},
	ToggleAll: Binding{Keys: []string{"a"}, Help: "a toggle all"},
}

var DefaultSearch = Search{
	Open:       Binding{Keys: []string{"enter"}, Help: "enter open"},
	DeleteChar: Binding{Keys: []string{"backspace"}, Help: ""},
	List:       DefaultList,
}

var DefaultHome = Home{
	Left:  Binding{Keys: []string{"h", "left", "k", "up"}, Help: ""},
	Right: Binding{Keys: []string{"l", "right", "j", "down"}, Help: ""},

	FlashCards:    Binding{Keys: []string{"f"}, Help: ""},
	TypingQuiz:    Binding{Keys: []string{"t"}, Help: ""},
	Statistics:    Binding{Keys: []string{"i"}, Help: ""},
	Search:        Binding{Keys: []string{"s"}, Help: ""},
	Configuration: Binding{Keys: []string{"c"}, Help: ""},
	DeckSelect:    Binding{Keys: []string{"ctrl+f"}, Help: ""},
	Palette:       Binding{Keys: []string{"p"}, Help: ""},
}

// DefaultRegistry is the default keybinding configuration for the application.
var DefaultRegistry = Registry{
	Global:     DefaultGlobal,
	List:       DefaultList,
	Home:       DefaultHome,
	Quiz:       DefaultQuiz,
	TypingQuiz: DefaultTypingQuiz,
	Decks:      DefaultDecks,
	Search:     DefaultSearch,
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
	Home *struct {
		Left  *BindingOverride `yaml:"left,omitempty"`
		Right *BindingOverride `yaml:"right,omitempty"`

		FlashCards    *BindingOverride `yaml:"flash_cards,omitempty"`
		TypingQuiz    *BindingOverride `yaml:"typing_quiz,omitempty"`
		Statistics    *BindingOverride `yaml:"statistics,omitempty"`
		Search        *BindingOverride `yaml:"search,omitempty"`
		Configuration *BindingOverride `yaml:"configuration,omitempty"`
		DeckSelect    *BindingOverride `yaml:"deck_select,omitempty"`
		Palette       *BindingOverride `yaml:"palette,omitempty"`
	} `yaml:"home,omitempty"`
	Quiz *struct {
		Reveal  *BindingOverride `yaml:"reveal,omitempty"`
		Again   *BindingOverride `yaml:"again,omitempty"`
		Hard    *BindingOverride `yaml:"hard,omitempty"`
		Good    *BindingOverride `yaml:"good,omitempty"`
		Easy    *BindingOverride `yaml:"easy,omitempty"`
		Inverse *BindingOverride `yaml:"inverse,omitempty"`
	} `yaml:"quiz,omitempty"`
	TypingQuiz *struct {
		Submit       *BindingOverride `yaml:"submit,omitempty"`
		Reveal       *BindingOverride `yaml:"reveal,omitempty"`
		Inverse      *BindingOverride `yaml:"inverse,omitempty"`
		PrevExample  *BindingOverride `yaml:"prev_example,omitempty"`
		NextExample  *BindingOverride `yaml:"next_example,omitempty"`
	} `yaml:"typing_quiz,omitempty"`
	Decks *struct {
		Toggle    *BindingOverride `yaml:"toggle,omitempty"`
		ToggleAll *BindingOverride `yaml:"toggle_all,omitempty"`
	} `yaml:"decks,omitempty"`
	Search *struct {
		Open       *BindingOverride `yaml:"open,omitempty"`
		DeleteChar *BindingOverride `yaml:"delete_char,omitempty"`
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
	if cfg.Home != nil {
		if cfg.Home.Left != nil {
			DefaultHome.Left = applyOverride(DefaultHome.Left, *cfg.Home.Left)
		}
		if cfg.Home.Right != nil {
			DefaultHome.Right = applyOverride(DefaultHome.Right, *cfg.Home.Right)
		}
		if cfg.Home.FlashCards != nil {
			DefaultHome.FlashCards = applyOverride(DefaultHome.FlashCards, *cfg.Home.FlashCards)
		}
		if cfg.Home.TypingQuiz != nil {
			DefaultHome.TypingQuiz = applyOverride(DefaultHome.TypingQuiz, *cfg.Home.TypingQuiz)
		}
		if cfg.Home.Statistics != nil {
			DefaultHome.Statistics = applyOverride(DefaultHome.Statistics, *cfg.Home.Statistics)
		}
		if cfg.Home.Search != nil {
			DefaultHome.Search = applyOverride(DefaultHome.Search, *cfg.Home.Search)
		}
		if cfg.Home.Configuration != nil {
			DefaultHome.Configuration = applyOverride(DefaultHome.Configuration, *cfg.Home.Configuration)
		}
		if cfg.Home.DeckSelect != nil {
			DefaultHome.DeckSelect = applyOverride(DefaultHome.DeckSelect, *cfg.Home.DeckSelect)
		}
		if cfg.Home.Palette != nil {
			DefaultHome.Palette = applyOverride(DefaultHome.Palette, *cfg.Home.Palette)
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
		if cfg.Quiz.Inverse != nil {
			DefaultQuiz.Inverse = applyOverride(DefaultQuiz.Inverse, *cfg.Quiz.Inverse)
		}
	}
	if cfg.TypingQuiz != nil {
		if cfg.TypingQuiz.Submit != nil {
			DefaultTypingQuiz.Submit = applyOverride(DefaultTypingQuiz.Submit, *cfg.TypingQuiz.Submit)
		}
		if cfg.TypingQuiz.Reveal != nil {
			DefaultTypingQuiz.Reveal = applyOverride(DefaultTypingQuiz.Reveal, *cfg.TypingQuiz.Reveal)
		}
		if cfg.TypingQuiz.Inverse != nil {
			DefaultTypingQuiz.Inverse = applyOverride(DefaultTypingQuiz.Inverse, *cfg.TypingQuiz.Inverse)
		}
		if cfg.TypingQuiz.PrevExample != nil {
			DefaultTypingQuiz.PrevExample = applyOverride(DefaultTypingQuiz.PrevExample, *cfg.TypingQuiz.PrevExample)
		}
		if cfg.TypingQuiz.NextExample != nil {
			DefaultTypingQuiz.NextExample = applyOverride(DefaultTypingQuiz.NextExample, *cfg.TypingQuiz.NextExample)
		}
	}
	if cfg.Decks != nil {
		if cfg.Decks.Toggle != nil {
			DefaultDecks.Toggle = applyOverride(DefaultDecks.Toggle, *cfg.Decks.Toggle)
		}
		if cfg.Decks.ToggleAll != nil {
			DefaultDecks.ToggleAll = applyOverride(DefaultDecks.ToggleAll, *cfg.Decks.ToggleAll)
		}
	}
	if cfg.Search != nil {
		if cfg.Search.Open != nil {
			DefaultSearch.Open = applyOverride(DefaultSearch.Open, *cfg.Search.Open)
		}
		if cfg.Search.DeleteChar != nil {
			DefaultSearch.DeleteChar = applyOverride(DefaultSearch.DeleteChar, *cfg.Search.DeleteChar)
		}
	}
	// Keep DefaultRegistry in sync with the per-group defaults.
	DefaultRegistry = Registry{
		Global:     DefaultGlobal,
		List:       DefaultList,
		Home:       DefaultHome,
		Quiz:       DefaultQuiz,
		TypingQuiz: DefaultTypingQuiz,
		Decks:      DefaultDecks,
		Search:     DefaultSearch,
	}
}
