package app

import (
	"strings"

	"crds/internal/ui"
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
	return ui.Theme.Primary.Render("Help Overlay") +
		"\n\n" +
		ui.Theme.Muted.Render("Press esc to close")
}

