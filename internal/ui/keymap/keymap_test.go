package keymap

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBindingMatch(t *testing.T) {
	b := Binding{Keys: []string{"up", "k"}, Help: "up"}

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"up key", "up", true},
		{"vim k", "k", true},
		{"unrelated", "enter", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Match(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune(tt.key)})); got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestBindingMatchCtrlC(t *testing.T) {
	b := Binding{Keys: []string{"ctrl+c"}, Help: "quit"}
	if !b.Match(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlC})) {
		t.Error("Match(ctrl+c) should be true")
	}
}

func TestBindingListHelp(t *testing.T) {
	tests := []struct {
		name string
		list BindingList
		want string
	}{
		{"single", BindingList{{Help: "a"}}, "a"},
		{"multiple", BindingList{{Help: "a"}, {Help: "b"}}, "a · b"},
		{"empty help skipped", BindingList{{Help: ""}, {Help: "b"}}, "b"},
		{"all empty", BindingList{{Help: ""}}, ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.list.Help(); got != tt.want {
				t.Errorf("Help() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultGlobal(t *testing.T) {
	if !DefaultGlobal.Quit.Match(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlC})) {
		t.Error("Global.Quit should match ctrl+c")
	}
	if !DefaultGlobal.Help.Match(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'?'}})) {
		t.Error("Global.Help should match ?")
	}
	if !DefaultGlobal.Back.Match(tea.KeyMsg(tea.Key{Type: tea.KeyEscape})) {
		t.Error("Global.Back should match esc")
	}
}

func TestDefaultGlobalFooter(t *testing.T) {
	footer := DefaultGlobal.Footer()
	if footer == "" {
		t.Error("Global.Footer() should not be empty")
	}
}

func TestDefaultListKeys(t *testing.T) {
	if !DefaultList.Up.Match(tea.KeyMsg(tea.Key{Type: tea.KeyUp})) {
		t.Error("List.Up should match up arrow")
	}
	if !DefaultList.Up.Match(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'k'}})) {
		t.Error("List.Up should match k")
	}
	if !DefaultList.Down.Match(tea.KeyMsg(tea.Key{Type: tea.KeyDown})) {
		t.Error("List.Down should match down arrow")
	}
	if !DefaultList.Down.Match(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}})) {
		t.Error("List.Down should match j")
	}
	if !DefaultList.Select.Match(tea.KeyMsg(tea.Key{Type: tea.KeyEnter})) {
		t.Error("List.Select should match enter")
	}
}

func TestDefaultListFooter(t *testing.T) {
	footer := DefaultList.Footer()
	if footer == "" {
		t.Error("List.Footer() should not be empty")
	}
}

func TestDefaultQuizKeys(t *testing.T) {
	if !DefaultQuiz.Reveal.Match(tea.KeyMsg(tea.Key{Type: tea.KeyEnter})) {
		t.Error("Quiz.Reveal should match enter")
	}
	grades := []struct {
		b Binding
		n string
	}{
		{DefaultQuiz.Again, "1"},
		{DefaultQuiz.Hard, "2"},
		{DefaultQuiz.Good, "3"},
		{DefaultQuiz.Easy, "4"},
	}
	for _, g := range grades {
		t.Run(g.n, func(t *testing.T) {
			if !g.b.Match(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune(g.n)})) {
				t.Errorf("should match key %q", g.n)
			}
		})
	}
}

func TestDefaultQuizFooter(t *testing.T) {
	if DefaultQuiz.Unrevealed() == "" {
		t.Error("Quiz.Unrevealed() should not be empty")
	}
	if DefaultQuiz.Revealed() == "" {
		t.Error("Quiz.Revealed() should not be empty")
	}
}

func TestDefaultSearchKeys(t *testing.T) {
	if !DefaultSearch.Open.Match(tea.KeyMsg(tea.Key{Type: tea.KeyEnter})) {
		t.Error("Search.Open should match enter")
	}
	if !DefaultSearch.DeleteChar.Match(tea.KeyMsg(tea.Key{Type: tea.KeyBackspace})) {
		t.Error("Search.DeleteChar should match backspace")
	}
}

func TestDefaultSearchFooter(t *testing.T) {
	footer := DefaultSearch.Footer()
	if footer == "" {
		t.Error("Search.Footer() should not be empty")
	}
}

func TestDefaultRegistryBindings(t *testing.T) {
	all := DefaultRegistry.Bindings()
	if len(all) == 0 {
		t.Fatal("Bindings() should not be empty")
	}
	groups := map[string]int{}
	for _, nb := range all {
		groups[nb.Group]++
	}
	if groups["Global"] == 0 {
		t.Error("missing Global bindings")
	}
	if groups["List"] == 0 {
		t.Error("missing List bindings")
	}
	if groups["Quiz"] == 0 {
		t.Error("missing Quiz bindings")
	}
	if groups["Search"] == 0 {
		t.Error("missing Search bindings")
	}
}

func TestApplyDefaultOverrides(t *testing.T) {
	// Save originals to restore after test.
	saveGlobal, saveList, saveQuiz, saveSearch := DefaultGlobal, DefaultList, DefaultQuiz, DefaultSearch
	defer func() {
		DefaultGlobal, DefaultList, DefaultQuiz, DefaultSearch = saveGlobal, saveList, saveQuiz, saveSearch
		DefaultRegistry = Registry{Global: saveGlobal, List: saveList, Quiz: saveQuiz, Search: saveSearch}
	}()

	helpStr := "? help"
	cfg := KeymapConfig{
		Global: &struct {
			Quit *BindingOverride `yaml:"quit,omitempty"`
			Help *BindingOverride `yaml:"help,omitempty"`
			Back *BindingOverride `yaml:"back,omitempty"`
		}{
			Quit: &BindingOverride{Keys: []string{"ctrl+q"}},
			Help: &BindingOverride{Help: &helpStr},
		},
		List: &struct {
			Up     *BindingOverride `yaml:"up,omitempty"`
			Down   *BindingOverride `yaml:"down,omitempty"`
			Select *BindingOverride `yaml:"select,omitempty"`
		}{
			Up: &BindingOverride{Keys: []string{"w"}},
		},
	}

	ApplyDefaultOverrides(cfg)

	// Verify overrides applied
	if len(DefaultGlobal.Quit.Keys) != 1 || DefaultGlobal.Quit.Keys[0] != "ctrl+q" {
		t.Errorf("Global.Quit.Keys = %v, want [ctrl+q]", DefaultGlobal.Quit.Keys)
	}
	if DefaultGlobal.Help.Help != "? help" {
		t.Errorf("Global.Help.Help = %q, want ? help", DefaultGlobal.Help.Help)
	}
	if len(DefaultGlobal.Help.Keys) != 1 || DefaultGlobal.Help.Keys[0] != "?" {
		t.Errorf("Global.Help.Keys should keep default, got %v", DefaultGlobal.Help.Keys)
	}
	if len(DefaultList.Up.Keys) != 1 || DefaultList.Up.Keys[0] != "w" {
		t.Errorf("List.Up.Keys = %v, want [w]", DefaultList.Up.Keys)
	}
	if len(DefaultList.Down.Keys) != 2 || DefaultList.Down.Keys[0] != "down" {
		t.Errorf("List.Down.Keys should keep default, got %v", DefaultList.Down.Keys)
	}

	// Verify Registry sync
	if DefaultRegistry.Global.Quit.Keys[0] != "ctrl+q" {
		t.Error("DefaultRegistry not in sync after ApplyDefaultOverrides")
	}
}

func TestDefaultRegistryFindBinding(t *testing.T) {
	tests := []struct {
		key      string
		wantNil  bool
		wantGrp  string
		wantAct  string
	}{
		{"ctrl+c", false, "Global", "Quit"},
		{"?", false, "Global", "Help"},
		{"esc", false, "Global", "Back"},
		{"up", false, "List", "Up"},
		{"k", false, "List", "Up"},
		{"down", false, "List", "Down"},
		{"j", false, "List", "Down"},
		{"enter", false, "List", "Select"},
		{"1", false, "Quiz", "Again"},
		{"2", false, "Quiz", "Hard"},
		{"3", false, "Quiz", "Good"},
		{"4", false, "Quiz", "Easy"},
		{"backspace", false, "Search", "DeleteChar"},
		{"nonexistent", true, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			nb := DefaultRegistry.FindBinding(tt.key)
			if tt.wantNil {
				if nb != nil {
					t.Errorf("FindBinding(%q) = %+v, want nil", tt.key, nb)
				}
				return
			}
			if nb == nil {
				t.Fatalf("FindBinding(%q) = nil, want %s/%s", tt.key, tt.wantGrp, tt.wantAct)
			}
			if nb.Group != tt.wantGrp {
				t.Errorf("Group = %q, want %q", nb.Group, tt.wantGrp)
			}
			if nb.Action != tt.wantAct {
				t.Errorf("Action = %q, want %q", nb.Action, tt.wantAct)
			}
		})
	}
}
