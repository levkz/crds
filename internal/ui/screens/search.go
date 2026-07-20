package screens

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/styles"
)

type searchEntry struct {
	ID    string
	front string
	back  []string
	notes string
}

type SearchModel struct {
	query   string
	cursor  int
	results []searchEntry
	cards   []ui.CardData
	focused bool
	width   int
	height  int
}

func NewSearch() *SearchModel {
	return &SearchModel{
		focused: true,
		width:   60,
		height:  24,
	}
}

func (m *SearchModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m SearchModel) Init() tea.Cmd { return nil }

func (m *SearchModel) SetSearchData(cards []ui.CardData) {
	m.cards = cards
	m.query = ""
	m.results = nil
	m.cursor = 0
}

func (m *SearchModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case keymap.DefaultList.Up.Match(msg):
			if m.cursor > 0 {
				m.cursor--
			}
		case keymap.DefaultList.Down.Match(msg):
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case keymap.DefaultSearch.Open.Match(msg):
			if len(m.results) > 0 {
				entry := m.results[m.cursor]
				return m, func() tea.Msg {
					return ui.NavigateToDetailMsg{
						Screen: ui.DetailScreen,
						Entry: ui.CardData{
							ID:    entry.ID,
							Front: entry.front,
							Back:  entry.back,
							Notes: entry.notes,
						},
					}
				}
			}
		case keymap.DefaultSearch.DeleteChar.Match(msg):
			if len(m.query) > 0 {
				runes := []rune(m.query)
				m.query = string(runes[:len(runes)-1])
				m.filterResults()
				m.cursor = 0
			}
		case keymap.DefaultSearch.FocusToggle.Match(msg):
			m.focused = !m.focused
		default:
			if m.focused && isPrintable(msg.String()) {
				m.query += msg.String()
				m.filterResults()
				m.cursor = 0
			}
		}
	}
	return m, nil
}

func (m *SearchModel) filterResults() {
	if m.query == "" || len(m.cards) == 0 {
		m.results = nil
		return
	}

	q := strings.ToLower(m.query)
	for _, card := range m.cards {
		if strings.Contains(strings.ToLower(card.Front), q) {
			m.results = append(m.results, searchEntry{
				ID:    card.ID,
				front: card.Front,
				back:  card.Back,
				notes: card.Notes,
			})
			continue
		}
		for _, t := range card.Back {
			if strings.Contains(strings.ToLower(t), q) {
				m.results = append(m.results, searchEntry{
					ID:    card.ID,
					front: card.Front,
					back:  card.Back,
					notes: card.Notes,
				})
				break
			}
		}
	}
}

func isPrintable(s string) bool {
	if len(s) != 1 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsPrint(r) && !unicode.IsControl(r)
}

func (m SearchModel) View() string {
	input := styles.FocusedInput().Render(m.query + "█")
	if !m.focused {
		input = styles.FocusedInput().Render(m.query)
	}

	var results string
	switch {
	case len(m.results) > 0:
		items := make([]string, len(m.results))
		for i, r := range m.results {
			items[i] = r.front + " → " + strings.Join(r.back, ", ")
		}
		results = components.RenderList(items, m.cursor, m.width)
	case m.query != "":
		results = styles.MutedText().Render("No results found")
	default:
		results = styles.MutedText().Render("Type to search vocabulary")
	}

	return layout.Page(
		components.Header("Search", m.width),
		layout.Column(input, results),
		components.Footer(keymap.DefaultSearch.Footer()+" · "+keymap.DefaultGlobal.Back.Help, m.width),
	)
}
