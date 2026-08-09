package screens

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type searchMode int

const (
	searchInput   searchMode = iota
	searchResults
)

type searchEntry struct {
	ID    string
	front string
	back  []string
	notes string
}

type SearchModel struct {
	query        string
	cursor       int
	scrollOffset int
	results      []searchEntry
	cards        []ui.CardData
	mode         searchMode
	width        int
	height       int
}

func NewSearch() *SearchModel {
	return &SearchModel{}
}

func (m *SearchModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m SearchModel) Init() tea.Cmd { return nil }

func (m *SearchModel) OnEnter() tea.Cmd {
	m.mode = searchInput
	return nil
}

func (m *SearchModel) OnLeave() tea.Cmd {
	m.query = ""
	m.results = nil
	m.cursor = 0
	m.scrollOffset = 0
	m.mode = searchInput
	return nil
}

func (m *SearchModel) SyncState(s ui.AppState) tea.Cmd {
	if s.Deck == nil {
		m.cards = nil
	} else {
		m.cards = s.Deck.Cards
	}
	if m.query != "" {
		m.filterResults()
		m.cursor = 0
	}
	return nil
}

func (m *SearchModel) topPadding() int {
	if m.height < 10 {
		return 0
	}
	n := m.height/4 - 3
	if n < 0 {
		return 0
	}
	return n
}

func (m *SearchModel) maxVisible() int {
	n := m.topPadding()
	avail := m.height - n - 9
	if avail < 1 {
		return 1
	}
	return avail
}

func (m *SearchModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case searchInput:
			return m.updateInput(msg)
		case searchResults:
			return m.updateResults(msg)
		}
	}
	return m, nil
}

func (m *SearchModel) updateInput(msg tea.KeyMsg) (ui.Screen, tea.Cmd) {
	switch {
	case keymap.DefaultList.Up.Match(msg),
		keymap.DefaultList.Down.Match(msg):
		if isPrintable(msg.String()) {
			m.query += msg.String()
			m.filterResults()
			m.cursor = 0
		}
	case keymap.DefaultSearch.Open.Match(msg):
		if len(m.results) > 0 {
			m.mode = searchResults
		}
	case keymap.DefaultSearch.DeleteChar.Match(msg):
		if len(m.query) > 0 {
			runes := []rune(m.query)
			m.query = string(runes[:len(runes)-1])
			m.filterResults()
			m.cursor = 0
		}
	default:
		if isPrintable(msg.String()) {
			m.query += msg.String()
			m.filterResults()
			m.cursor = 0
		}
	}
	return m, nil
}

func (m *SearchModel) updateResults(msg tea.KeyMsg) (ui.Screen, tea.Cmd) {
	switch {
	case keymap.DefaultList.Up.Match(msg):
		if m.cursor > 0 {
			m.cursor--
			m.adjustScroll()
		}
	case keymap.DefaultList.Down.Match(msg):
		if m.cursor < len(m.results)-1 {
			m.cursor++
			m.adjustScroll()
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
	}
	return m, nil
}

func (m *SearchModel) HandleBack() bool {
	if m.mode == searchResults {
		m.mode = searchInput
		return true
	}
	return false
}

func (m *SearchModel) adjustScroll() {
	avail := m.maxVisible()
	if avail <= 2 {
		m.scrollOffset = m.cursor
		return
	}
	visible := avail - 2 // worst-case: both indicator lines shown
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}
}

func (m *SearchModel) filterResults() {
	m.results = filterCards(m.cards, m.query)
	m.scrollOffset = 0
}

func isPrintable(s string) bool {
	if len(s) != 1 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsPrint(r) && !unicode.IsControl(r)
}

type highlightSegment struct {
	text        string
	highlighted bool
}

func splitHighlight(text, query string) []highlightSegment {
	if query == "" {
		return []highlightSegment{{text: text}}
	}
	lower := strings.ToLower(text)
	q := strings.ToLower(query)
	var segments []highlightSegment
	pos := 0
	for {
		idx := strings.Index(lower[pos:], q)
		if idx == -1 {
			segments = append(segments, highlightSegment{text: text[pos:]})
			break
		}
		if idx > 0 {
			segments = append(segments, highlightSegment{text: text[pos : pos+idx]})
		}
		end := pos + idx + len(q)
		segments = append(segments, highlightSegment{text: text[pos+idx : end], highlighted: true})
		pos = end
	}
	return segments
}

func (m *SearchModel) renderInput() string {
	if m.mode == searchInput {
		return ui.Theme.Background.Render(m.query + "█")
	}
	return ui.Theme.Background.Render(m.query)
}

func (m *SearchModel) renderCenteredResults() string {
	switch {
	case len(m.results) > 0:
		maxItemWidth := m.width - 3
		if maxItemWidth < 1 {
			maxItemWidth = 1
		}
		maxVisible := m.maxVisible()

		items := make([]string, len(m.results))
		for i, r := range m.results {
			items[i] = r.front + " → " + strings.Join(r.back, ", ")
		}

		sel := -1
		offset := 0
		if m.mode == searchResults {
			sel = m.cursor
			offset = m.scrollOffset
		}

		visible := items
		relSel := sel
		showAbove := false
		showBelow := false

		if maxVisible > 0 && len(items) > maxVisible {
			showAbove = offset > 0
			itemLimit := maxVisible
			if showAbove {
				itemLimit--
				if itemLimit < 1 {
					itemLimit = 1
					showAbove = false
				}
			}
			end := offset + itemLimit
			if end > len(items) {
				end = len(items)
			}
			showBelow = end < len(items)
			if showBelow {
				itemLimit--
				if itemLimit < 1 {
					itemLimit = 1
					showBelow = false
				}
				end = offset + itemLimit
				if end > len(items) {
					end = len(items)
				}
				showBelow = end < len(items)
			}
			visible = items[offset:end]
			relSel = sel - offset
		}

		bg := ui.Theme.Palette.Selection
		selNormal := ui.Theme.Primary.Background(bg)
		selHighlight := ui.Theme.Secondary.Background(bg)

		var b strings.Builder
		first := true

		if showAbove {
			if !first {
				b.WriteString("\n")
			}
			first = false
			b.WriteString(layout.Center(styles.MutedText().Render("  ↑ more above"), m.width))
		}

		for i, item := range visible {
			if !first {
				b.WriteString("\n")
			}
			first = false

			truncated := renderer.Truncate(item, maxItemWidth)
			selected := i == relSel

			var line strings.Builder
			if selected {
				line.WriteString(selNormal.Render(ui.Theme.Icons.Navigate + " "))
			} else {
				line.WriteString(styles.MutedText().Render("  "))
			}

			if m.query != "" {
				for _, seg := range splitHighlight(truncated, m.query) {
					if seg.highlighted {
						if selected {
							line.WriteString(selHighlight.Render(seg.text))
						} else {
							line.WriteString(ui.Theme.Secondary.Render(seg.text))
						}
					} else {
						if selected {
							line.WriteString(selNormal.Render(seg.text))
						} else {
							line.WriteString(styles.MutedText().Render(seg.text))
						}
					}
				}
			} else {
				if selected {
					line.WriteString(selNormal.Render(truncated))
				} else {
					line.WriteString(styles.MutedText().Render(truncated))
				}
			}

			b.WriteString(layout.Center(line.String(), m.width))
		}

		if showBelow {
			if !first {
				b.WriteString("\n")
			}
			b.WriteString(layout.Center(styles.MutedText().Render("  ↓ more below"), m.width))
		}

		return b.String()

	case m.query != "":
		return layout.Center(styles.MutedText().Render("No results found"), m.width)
	default:
		return layout.Center(styles.MutedText().Render("Type to search vocabulary"), m.width)
	}
}

func (m SearchModel) View() string {
	var footer string
	if m.mode == searchInput {
		footer = "enter search · " + keymap.DefaultGlobal.Back.Help
	} else {
		footer = keymap.DefaultSearch.Footer() + " · " + keymap.DefaultGlobal.Back.Help
	}

	return layout.Page(
		ui.Theme.Muted.Render("search"),
		layout.VSpace(m.topPadding())+
			layout.Center(m.renderInput(), m.width)+"\n\n"+
			m.renderCenteredResults(),
		components.Footer(footer, m.width),
		m.height,
	)
}
