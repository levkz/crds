package screens

import (
	"slices"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/layout"
	"crds/internal/ui/renderer"
)

type columnState struct {
	items        []string
	selected     map[string]bool
	cursor       int
	scrollOffset int
	searchQuery  string
	searchActive bool
}

func newColumn(items []string, selected []string) columnState {
	sel := make(map[string]bool, len(selected))
	for _, s := range selected {
		sel[s] = true
	}
	return columnState{
		items:    items,
		selected: sel,
	}
}

type DeckSelectModel struct {
	decks     columnState
	tags      columnState
	activeCol int // 0=decks, 1=tags
	width     int
	height    int
	deckTags  map[string][]string // deckID → tags for cross-filtering
}

func NewDeckSelect() *DeckSelectModel {
	return &DeckSelectModel{
		activeCol: 0,
	}
}

func (m *DeckSelectModel) SyncState(s ui.AppState) tea.Cmd {
	m.decks = newColumn(s.AllDecks, s.SelectedDecks)
	m.tags = newColumn(s.AllTags, s.SelectedTags)
	m.deckTags = s.AllDeckTags
	return nil
}

func (m *DeckSelectModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *DeckSelectModel) Init() tea.Cmd { return nil }

// --- Screen lifecycle ---

func (m *DeckSelectModel) OnEnter() tea.Cmd { return nil }

func (m *DeckSelectModel) OnLeave() tea.Cmd {
	return func() tea.Msg {
		return ui.DeckSelectionChangedMsg{
			Selected:     m.selectedNames(m.decks),
			SelectedTags: m.selectedNames(m.tags),
		}
	}
}

func (m *DeckSelectModel) selectedNames(col columnState) []string {
	var out []string
	for _, name := range col.items {
		if col.selected[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// --- Column filtering ---

func (m *DeckSelectModel) filteredDecks() []string {
	all := m.decks.items
	selectedTags := m.selectedTagList()

	// Filter by search query
	if m.decks.searchQuery != "" {
		q := strings.ToLower(m.decks.searchQuery)
		var filtered []string
		for _, d := range all {
			if m.decks.selected[d] || strings.Contains(strings.ToLower(d), q) {
				filtered = append(filtered, d)
				continue
			}
			// Filter by selected tags (OR: deck has ANY selected tag)
			if len(selectedTags) > 0 {
				for _, t := range selectedTags {
					if m.deckHasTag(d, t) {
						filtered = append(filtered, d)
						break
					}
				}
			}
		}
		all = filtered
	}

	return all
}

func (m *DeckSelectModel) filteredTags() []string {
	all := m.tags.items

	// Filter by search query
	if m.tags.searchQuery != "" {
		q := strings.ToLower(m.tags.searchQuery)
		var filtered []string
		for _, t := range all {
			if strings.Contains(strings.ToLower(t), q) {
				filtered = append(filtered, t)
			}
		}
		all = filtered
	}

	// Filter by selected decks (OR: tag in ANY selected deck)
	selectedDecks := m.selectedDeckList()
	selectedTags := m.selectedTagList()
	if len(selectedDecks) > 0 {
		var filtered []string
		for _, t := range all {
			if slices.Contains(selectedTags, t) {
				filtered = append(filtered, t)
				continue
			}
			for _, d := range selectedDecks {
				if m.deckHasTag(d, t) {
					filtered = append(filtered, t)
					break
				}
			}
		}
		all = filtered
	}

	return all
}

func (m *DeckSelectModel) selectedDeckList() []string {
	var out []string
	for _, d := range m.decks.items {
		if m.decks.selected[d] {
			out = append(out, d)
		}
	}
	return out
}

func (m *DeckSelectModel) selectedTagList() []string {
	var out []string
	for _, t := range m.tags.items {
		if m.tags.selected[t] {
			out = append(out, t)
		}
	}
	return out
}

func (m *DeckSelectModel) deckHasTag(deckID, tag string) bool {
	tags, ok := m.deckTags[deckID]
	if !ok {
		return false
	}
	return slices.Contains(tags, tag)
}

// --- Update ---

func (m *DeckSelectModel) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *DeckSelectModel) col() *columnState {
	if m.activeCol == 0 {
		return &m.decks
	}
	return &m.tags
}

func (m *DeckSelectModel) otherCol() *columnState {
	if m.activeCol == 0 {
		return &m.tags
	}
	return &m.decks
}

func (m *DeckSelectModel) handleKey(msg tea.KeyMsg) (ui.Screen, tea.Cmd) {
	col := m.col()

	// If search is active in this column, typing goes to search
	if col.searchActive {
		switch {
		case keymap.DefaultList.Select.Match(msg):
			col.searchActive = false
			return m, nil

		case msg.String() == "backspace":
			if len(col.searchQuery) > 0 {
				col.searchQuery = col.searchQuery[:len(col.searchQuery)-1]
			}
			m.resetCursor(col)
			return m, nil

		default:
			if isPrintable(msg.String()) {
				col.searchQuery += msg.String()
				m.resetCursor(col)
				return m, nil
			}
		}
	}

	// Normal key handling
	items := m.filteredItems(col)

	switch {
	case keymap.DefaultDeckSelect.Up.Match(msg):
		if col.cursor > 0 {
			col.cursor--
			m.adjustScroll(col)
		}

	case keymap.DefaultDeckSelect.Down.Match(msg):
		if col.cursor < len(items)-1 {
			col.cursor++
			m.adjustScroll(col)
		}

	case keymap.DefaultDeckSelect.Toggle.Match(msg):
		if len(items) > 0 && col.cursor < len(items) {
			name := items[col.cursor]
			col.selected[name] = !col.selected[name]
			m.resetCursor(m.otherCol())
		}

	case keymap.DefaultDeckSelect.ToggleAll.Match(msg):
		all := true
		for _, name := range items {
			if !col.selected[name] {
				all = false
				break
			}
		}
		for _, name := range items {
			col.selected[name] = !all
		}
		m.resetCursor(m.otherCol())

	case keymap.DefaultDeckSelect.SearchToggle.Match(msg):
		col.searchActive = true
		return m, nil

	case keymap.DefaultDeckSelect.NextColumn.Match(msg):
		m.activeCol = (m.activeCol + 1) % 2

	case keymap.DefaultDeckSelect.PrevColumn.Match(msg):
		m.activeCol = (m.activeCol - 1 + 2) % 2

	case keymap.DefaultDeckSelect.Select.Match(msg):
		if len(items) > 0 {
			return m, func() tea.Msg {
				return ui.NavigateToMsg{Screen: ui.HomeScreen}
			}
		}
	}

	return m, nil
}

func (m *DeckSelectModel) HandleBack() bool {
	col := m.col()
	if col.searchActive {
		if col.searchQuery != "" {
			col.searchQuery = ""
			m.resetCursor(col)
			return true
		}
		col.searchActive = false
	}
	return false
}

func (m *DeckSelectModel) filteredItems(col *columnState) []string {
	if col == &m.decks {
		return m.filteredDecks()
	}
	return m.filteredTags()
}

func (m *DeckSelectModel) resetCursor(col *columnState) {
	items := m.filteredItems(col)
	if col.cursor >= len(items) && len(items) > 0 {
		col.cursor = len(items) - 1
	}
	if col.cursor < 0 {
		col.cursor = 0
	}
	col.scrollOffset = 0
}

func (m *DeckSelectModel) maxVisible(_ *columnState) int {
	return m.height - 6
}

func (m *DeckSelectModel) adjustScroll(col *columnState) {
	avail := m.maxVisible(col)
	visible := avail - 2
	if visible < 1 {
		col.scrollOffset = col.cursor
		return
	}
	if col.cursor < col.scrollOffset {
		col.scrollOffset = col.cursor
	}
	if col.cursor >= col.scrollOffset+visible {
		col.scrollOffset = col.cursor - visible + 1
	}
}

func (m *DeckSelectModel) padLines(lines []string, width int) string {
	if len(lines) == 0 || width <= 0 {
		return ""
	}
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		vw := renderer.VisibleWidth(line)
		b.WriteString(line)
		if vw < width {
			b.WriteString(strings.Repeat(" ", width-vw))
		}
	}
	return b.String()
}

// --- View ---

func (m *DeckSelectModel) View() string {
	margin := 4
	contentW := m.width - 2*margin
	if contentW < 20 {
		contentW = 20
	}
	leftW := contentW * 45 / 100
	rightW := contentW - leftW - 1
	if rightW < 10 {
		rightW = 10
	}

	left := m.renderColumn("Decks", &m.decks, leftW)
	right := m.renderColumn("Tags", &m.tags, rightW)
	divider := ui.Theme.Secondary.Render("│")

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right)
	body = lipgloss.NewStyle().PaddingLeft(margin).PaddingRight(margin).Render(body)

	return layout.Page(
		components.Header("Deck Selection", m.width),
		body,
		components.Footer(keymap.DefaultDeckSelect.Footer(), m.width),
		m.height,
	)
}

func (m *DeckSelectModel) renderColumn(title string, col *columnState, width int) string {
	var lines []string

	// Column header
	isActive := (col == &m.decks && m.activeCol == 0) || (col == &m.tags && m.activeCol == 1)
	if isActive {
		lines = append(lines, ui.Theme.Header.Render(" "+title+" "))
	} else {
		lines = append(lines, ui.Theme.Muted.Render(" "+title+" "))
	}

	// Search input bar
	if col.searchActive {
		q := col.searchQuery
		cursor := " "
		searchLabel := ui.Theme.Primary.Render(" search: ")
		queryText := ui.Theme.Muted.Render(q)
		if q == "" {
			queryText = ui.Theme.Muted.Render("")
		}
		if col.searchActive {
			cursor = "█"
		}
		searchLine := searchLabel + queryText + ui.Theme.Muted.Render(cursor)
		lines = append(lines, renderer.Truncate(searchLine, width))
	} else {
		hint := " " + ui.Theme.Secondary.Render("[s]") + ui.Theme.Muted.Italic(true).Render("earch: "+col.searchQuery)
		lines = append(lines, hint)
	}

	// Items list
	items := m.filteredItems(col)
	avail := m.maxVisible(col)

	if len(items) == 0 {
		if width > 10 {
			lines = append(lines, ui.Theme.Muted.Render(" (none)"))
		}
		for len(lines) < 2+avail {
			lines = append(lines, "")
		}
		return m.padLines(lines, width)
	}

	// Always reserve 2 lines for scroll indicators
	showAbove := col.scrollOffset > 0
	if showAbove {
		lines = append(lines, ui.Theme.Muted.Render(" "+ui.Theme.Icons.ArrowUp+" more "))
	}

	itemLines := max(avail-2, 1)
	end := min(col.scrollOffset+itemLines, len(items))

	var cursorItemStyle *lipgloss.Style

	for i := col.scrollOffset; i < end; i++ {
		if col.selected[items[i]] {
			cursorItemStyle = &ui.Theme.Secondary
		} else {
			cursorItemStyle = &ui.Theme.Primary
		}

		if i == col.cursor {
			line := cursorItemStyle.Render(ui.Theme.Icons.Navigate + " " + items[i])
			lines = append(lines, renderer.Truncate(line, width))
		} else if col.selected[items[i]] {
			line := ui.Theme.Secondary.Render(ui.Theme.Icons.Check + " " + items[i])
			lines = append(lines, renderer.Truncate(line, width))
		} else {
			lines = append(lines, renderer.Truncate("  "+items[i], width))
		}
	}

	showBelow := end < len(items)
	if showBelow {
		lines = append(lines, ui.Theme.Muted.Render(" "+ui.Theme.Icons.ArrowDown+" more "))
	}

	// Pad to fixed height: header(1) + search(1) + avail(item area) = 2 + avail
	for len(lines) < 2+avail {
		lines = append(lines, "")
	}

	return m.padLines(lines, width)
}
