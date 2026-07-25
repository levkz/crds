package app

import (
	"fmt"
	"strings"

	"crds/internal/ui"
	components "crds/internal/ui/components/display"
	"crds/internal/ui/keymap"
	"crds/internal/ui/renderer"
	"github.com/charmbracelet/lipgloss"
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

	output := b.String()

	if m.Global.Notification != nil {
		lines := strings.Split(output, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			if renderer.StripANSI(strings.TrimSpace(lines[i])) != "" {
				if m.Width > 0 {
					visible := renderer.StripANSI(lines[i])
					trimmed := strings.TrimRight(visible, " ")
					lines[i] = components.StatusBar(
						trimmed,
						m.Global.Notification.Text,
						m.Width,
						ui.Theme.Palette.Surface,
					)
				}
				break
			}
		}
		output = strings.Join(lines, "\n")
	}

	return fillBackground(output, m.Width)
}

func fillBackground(s string, width int) string {
	if width < 1 || s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	bgStyle := lipgloss.NewStyle().Background(ui.Theme.Palette.Background)
	for i, line := range lines {
		vw := renderer.VisibleWidth(line)
		if vw < width {
			lines[i] = line + bgStyle.Render(strings.Repeat(" ", width-vw))
		}
	}
	return strings.Join(lines, "\n")
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

