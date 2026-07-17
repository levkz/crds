package screens

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/components"
	"crds/internal/ui/styles"
)

type SearchModel struct {
	query   string
	cursor  int
	results []string
	focused bool
}

func NewSearch() SearchModel {
	return SearchModel{
		focused: true,
	}
}

func (m SearchModel) Init() tea.Cmd { return nil }

func (m SearchModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.results) > 0 {
				return m, func() tea.Msg {
					return ui.NavigateToMsg{Screen: ui.DetailScreen}
				}
			}
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.results = filterResults(m.query)
				m.cursor = 0
			}
		case "tab":
			m.focused = !m.focused
		default:
			if m.focused && isPrintable(msg.String()) {
				m.query += msg.String()
				m.results = filterResults(m.query)
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func isPrintable(s string) bool {
	if len(s) != 1 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsPrint(r) && !unicode.IsControl(r)
}

func filterResults(query string) []string {
	if query == "" {
		return nil
	}
	return []string{"Results for: " + query}
}

func (m SearchModel) View() string {
	var b strings.Builder
	b.WriteString(components.Header("Search"))
	b.WriteString("\n\n")

	if m.focused {
		b.WriteString(styles.FocusedInput().Render(m.query + "█"))
	} else {
		b.WriteString(styles.FocusedInput().Render(m.query))
	}

	b.WriteString("\n\n")

	if len(m.results) > 0 {
		b.WriteString(components.RenderList(m.results, m.cursor))
	} else if m.query != "" {
		b.WriteString(styles.MutedText().Render("No results found"))
	} else {
		b.WriteString(styles.MutedText().Render("Type to search vocabulary"))
	}

	b.WriteString("\n\n")
	b.WriteString(components.Footer("type to search · ↑/↓ navigate · enter open · esc back"))
	return b.String()
}
