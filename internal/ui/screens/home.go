package screens

import (
	_ "embed"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
)

var logoASCII string

//go:embed logo.txt
var logoData string

func init() {
	logoASCII = strings.TrimRight(logoData, "\n")
}

type menuItem struct {
	name   string
	screen ui.ScreenIndex
	key    *keymap.Binding
}

type HomeModel struct {
	cursor    int
	menuItems []menuItem
	width     int
	height    int
}

func NewHome() *HomeModel {
	items := []menuItem{
		{"flash cards", ui.QuizScreen, &keymap.DefaultHome.FlashCards},
		{"typing quiz", ui.TypingQuizScreen, &keymap.DefaultHome.TypingQuiz},
		{"statistics", ui.StatisticsScreen, &keymap.DefaultHome.Statistics},
		{"search", ui.SearchScreen, &keymap.DefaultHome.Search},
		{"configuration", ui.SettingsScreen, &keymap.DefaultHome.Configuration},
		{"deck selection", ui.NoScreen, &keymap.DefaultHome.DeckSelect},
		{"theme palette", ui.PaletteScreen, &keymap.DefaultHome.Palette},
	}
	return &HomeModel{menuItems: items}
}

func (m *HomeModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m HomeModel) Init() tea.Cmd { return nil }

func (m *HomeModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultHome.Left.Match(msg):
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.menuItems) - 1
			}
		case keymap.DefaultHome.Right.Match(msg):
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case keymap.DefaultList.Select.Match(msg):
			return m.navigateTo(m.menuItems[m.cursor])
		default:
			for _, item := range m.menuItems {
				if item.key.Match(msg) {
					return m.navigateTo(item)
				}
			}
		}
	}
	return m, nil
}

func (m *HomeModel) navigateTo(item menuItem) (ui.Screen, tea.Cmd) {
	if item.screen == ui.NoScreen {
		return m, func() tea.Msg { return ui.ShowDeckSelectionMsg{} }
	}
	return m, func() tea.Msg { return ui.NavigateToMsg{Screen: item.screen} }
}

func (m HomeModel) View() string {
	logo := renderLogo(m.width)
	menu := renderMenu(m.menuItems, m.cursor, m.width)
	footer := components.Footer(keymap.DefaultGlobal.Help.Help, m.width)

	body := logo + "\n\n" + menu
	bodyHeight := strings.Count(body, "\n") + 1
	footerHeight := strings.Count(footer, "\n") + 1

	avail := m.height - bodyHeight - 1 - footerHeight
	if avail < 0 {
		avail = 0
	}
	topPad := avail / 2
	bottomPad := avail - topPad

	return strings.Repeat("\n", topPad) + body + strings.Repeat("\n", bottomPad) + "\n\n" + footer
}

func renderLogo(width int) string {
	return layout.Center(ui.Theme.Primary.Render(logoASCII), width)
}

func renderMenu(items []menuItem, cursor int, width int) string {
	rowSize := 3
	var rows []string
	for i := 0; i < len(items); i += rowSize {
		end := i + rowSize
		if end > len(items) {
			end = len(items)
		}
		var cells []string
		for j, item := range items[i:end] {
			idx := i + j
			cell := renderMenuItem(item, idx, cursor == idx)
			if j < end-i-1 {
				cell += "  "
			}
			cells = append(cells, cell)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		row = layout.Center(row, width)
		rows = append(rows, row)
	}
	return lipgloss.JoinVertical(lipgloss.Top, rows...)
}

func renderMenuItem(item menuItem, _ int, selected bool) string {
	primary := item.key.Keys[0]
	nameStyle := ui.Theme.Muted
	if selected {
		nameStyle = ui.Theme.Primary
	}

	if isSingleLetter(primary) {
		lowerPrimary := strings.ToLower(primary)
		if idx := strings.Index(strings.ToLower(item.name), lowerPrimary); idx >= 0 {
			before := item.name[:idx]
			letter := string(item.name[idx])
			if primary != strings.ToLower(primary) {
				letter = primary
			}
			after := item.name[idx+len(lowerPrimary):]
			return nameStyle.Render(before) +
				ui.Theme.Accent.Render("["+letter+"]") +
				nameStyle.Render(after)
		}
	}

	display := shortcutDisplay(primary)
	return nameStyle.Render(item.name) + " " + ui.Theme.Accent.Render("["+display+"]")
}

func isSingleLetter(key string) bool {
	return len(key) == 1 && ((key[0] >= 'a' && key[0] <= 'z') || (key[0] >= 'A' && key[0] <= 'Z'))
}

func shortcutDisplay(key string) string {
	if isSingleLetter(key) {
		return strings.ToUpper(key)
	}
	switch {
	case strings.HasPrefix(key, "ctrl+"):
		return "C-" + strings.ToUpper(strings.TrimPrefix(key, "ctrl+"))
	case strings.HasPrefix(key, "alt+"):
		return "A-" + strings.ToUpper(strings.TrimPrefix(key, "alt+"))
	case strings.HasPrefix(key, "shift+"):
		return "S-" + strings.ToUpper(strings.TrimPrefix(key, "shift+"))
	default:
		return strings.ToUpper(key)
	}
}
