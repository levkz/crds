package components

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

type TreeSelectMsg struct {
	NodeID string
}

type TreeNode struct {
	Label    string
	ID       string
	Children []TreeNode
}

type TreeModel struct {
	cursor   int
	expanded map[string]bool
	focused  bool
	keys     NavigationKeys
}

func NewTree(keys ...NavigationKeys) TreeModel {
	k := DefaultNavigationKeys
	if len(keys) > 0 {
		k = keys[0]
	}
	return TreeModel{
		expanded: make(map[string]bool),
		keys:     k,
	}
}

func (m TreeModel) Init() tea.Cmd { return nil }

func (m TreeModel) Update(msg tea.Msg) (TreeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch {
		case keyIn(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case keyIn(msg, m.keys.Down):
			m.cursor++
		case keyIn(msg, m.keys.Home):
			m.cursor = 0
		case keyIn(msg, m.keys.End):
			m.cursor = math.MaxInt
		case keyIn(msg, m.keys.Confirm):
			return m, func() tea.Msg {
				return TreeSelectMsg{NodeID: ""}
			}
		}
	}
	return m, nil
}

func (m TreeModel) View(nodes []TreeNode, width int) string {
	if len(nodes) == 0 {
		return ""
	}

	vis := flattenNodes(nodes, m.expanded, 0)
	if len(vis) == 0 {
		return ""
	}

	cursor := m.cursor
	if cursor >= len(vis) {
		cursor = len(vis) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	var b strings.Builder
	for i, vn := range vis {
		if i > 0 {
			b.WriteString("\n")
		}

		indent := strings.Repeat("  ", vn.depth)
		hasChildren := len(vn.node.Children) > 0
		isExpanded := m.expanded[vn.node.ID]

		icon := " "
		switch {
		case hasChildren && isExpanded:
			icon = "▼"
		case hasChildren && !isExpanded:
			icon = "▶"
		}

		available := width - renderer.VisibleWidth(indent) - 2
		if available < 1 {
			available = 1
		}
		truncated := renderer.Truncate(vn.node.Label, available)

		line := indent + icon + " " + truncated

		if i == cursor && m.focused {
			b.WriteString(styles.SelectedItem().Render(
				ui.Theme.Icons.Navigate + " " + line,
			))
		} else {
			b.WriteString("  " + line)
		}
	}
	return b.String()
}

type visibleNode struct {
	node  TreeNode
	depth int
}

func flattenNodes(nodes []TreeNode, expanded map[string]bool, depth int) []visibleNode {
	var result []visibleNode
	for _, node := range nodes {
		result = append(result, visibleNode{node, depth})
		if expanded[node.ID] && len(node.Children) > 0 {
			result = append(result, flattenNodes(node.Children, expanded, depth+1)...)
		}
	}
	return result
}

func (m *TreeModel) ExpandAtCursor(nodes []TreeNode) {
	vis := flattenNodes(nodes, m.expanded, 0)
	if m.cursor < 0 || m.cursor >= len(vis) {
		return
	}
	id := vis[m.cursor].node.ID
	if len(vis[m.cursor].node.Children) > 0 {
		m.expanded[id] = true
	}
}

func (m *TreeModel) CollapseAtCursor(nodes []TreeNode) {
	vis := flattenNodes(nodes, m.expanded, 0)
	if m.cursor < 0 || m.cursor >= len(vis) {
		return
	}
	id := vis[m.cursor].node.ID
	if m.expanded[id] {
		delete(m.expanded, id)
	}
}

func (m *TreeModel) ExpandAll(nodes []TreeNode) {
	var walk func(n TreeNode)
	walk = func(n TreeNode) {
		for _, child := range n.Children {
			m.expanded[child.ID] = true
			walk(child)
		}
	}
	for _, n := range nodes {
		m.expanded[n.ID] = true
		walk(n)
	}
}

func (m *TreeModel) CollapseAll() {
	m.expanded = make(map[string]bool)
}

func (m *TreeModel) Cursor() int        { return m.cursor }
func (m *TreeModel) SetCursor(n int)    { m.cursor = n }
func (m *TreeModel) Focus()             { m.focused = true }
func (m *TreeModel) Blur()              { m.focused = false }
func (m *TreeModel) Focused() bool      { return m.focused }
func (m *TreeModel) IsExpanded(id string) bool { return m.expanded[id] }
