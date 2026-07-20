package app

import (
	"fmt"
	"strings"

	"crds/internal/ui"
	"crds/internal/ui/keymap"
	"crds/internal/ui/renderer"
)

func (m Model) View() string {
	var b strings.Builder

	if m.Global.Overlay != NoOverlay {
		b.WriteString(renderOverlay(m.Global.Overlay))
	} else {
		if screen, ok := m.Navigator.CurrentScreen(); ok {
			b.WriteString(screen.View())
		}
	}

	if m.Global.Notification != nil {
		b.WriteString("\n")
		b.WriteString(ui.Theme.Muted.Render(m.Global.Notification.Text))
	}

	return b.String()
}

func renderOverlay(t OverlayType) string {
	switch t {
	case HelpOverlay:
		return renderHelpOverlay()
	default:
		return ""
	}
}

func renderHelpOverlay() string {
	var b strings.Builder
	b.WriteString(ui.Theme.Primary.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	var maxKeyWidth int
	var groups []struct {
		name string
		bind []struct{ keys, help string }
	}
	var currentGroup string
	var currentBind []struct{ keys, help string }
	for _, nb := range keymap.DefaultRegistry.Bindings() {
		if nb.Group != currentGroup {
			if currentBind != nil {
				groups = append(groups, struct {
					name string
					bind []struct{ keys, help string }
				}{currentGroup, currentBind})
			}
			currentGroup = nb.Group
			currentBind = nil
		}
		keys := strings.Join(nb.Binding.Keys, "/")
		if w := renderer.VisibleWidth(keys); w > maxKeyWidth {
			maxKeyWidth = w
		}
		currentBind = append(currentBind, struct{ keys, help string }{keys, nb.Binding.Help})
	}
	if currentBind != nil {
		groups = append(groups, struct {
			name string
			bind []struct{ keys, help string }
		}{currentGroup, currentBind})
	}

	keyColWidth := maxKeyWidth + 2
	if keyColWidth < 4 {
		keyColWidth = 4
	}

	for i, group := range groups {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(ui.Theme.Muted.Render(group.name))
		b.WriteString("\n")
		for _, bind := range group.bind {
			b.WriteString(fmt.Sprintf("  %-*s  %s\n", keyColWidth, bind.keys, bind.help))
		}
	}

	b.WriteString("\n")
	b.WriteString(ui.Theme.Muted.Render("Press " + keymap.DefaultGlobal.Back.Help + " to close"))
	return b.String()
}

