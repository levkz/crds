package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/theme"
)

// dispatchEvent routes a message to its registered handler or the active screen.
func (m Model) dispatchEvent(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.dispatchKeyEvent(msg)

	case ShowOverlayMsg:
		return m.WithOverlay(msg.Type), nil

	case HideOverlayMsg:
		return m.WithoutOverlay(), nil

	case ShowNotificationMsg:
		return m.WithNotification(msg.Text), nil

	case HideNotificationMsg:
		return m.WithoutNotification(), nil

	case SetLoadingMsg:
		return m.WithLoading(msg.Loading), nil

	case ui.NavigateToMsg:
		return m.transitionTo(msg.Screen)

	case DataLoadedMsg:
		return m.WithoutNotification(), nil

	case DataErrorMsg:
		return m.WithNotification("Error loading "+msg.Kind.String()+": "+msg.Err.Error()), nil

	case SavedMsg:
		return m.WithNotification("Saved " + msg.Kind.String()), nil

	case SaveErrorMsg:
		return m.WithNotification("Error saving "+msg.Kind.String()+": "+msg.Err.Error()), nil

	case ConfigUpdatedMsg:
		m.Config = msg.Config
		return m, nil

	case ui.ThemeSwitchMsg:
		th, err := theme.Switch(msg.Name)
		if err != nil {
			return m.WithNotification("Theme not found: " + msg.Name), nil
		}
		ui.SetTheme(th)
		return m.WithNotification("Switched to " + msg.Name + " theme"), nil

	case TickMsg:
		return m, TickCmd()

	default:
		if m.Global.Overlay != NoOverlay {
			return m, nil
		}
		return m.forwardToScreen(msg)
	}
}

// dispatchKeyEvent handles global keyboard shortcuts.
func (m Model) dispatchKeyEvent(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case m.Config.KeyQuit:
		return m, tea.Sequence(m.ShutdownCmd(), tea.Quit)
	case m.Config.KeyHelp:
		return m.WithOverlay(HelpOverlay), nil
	case "esc":
		if m.Global.Overlay != NoOverlay {
			return m.WithoutOverlay(), nil
		}
		if m.Navigator.CanGoBack() {
			return m.popToPrevious()
		}
		return m.transitionTo(ui.HomeScreen)
	}
	return m.forwardToScreen(msg)
}

// forwardToScreen sends a message to the currently active screen model.
func (m Model) forwardToScreen(msg tea.Msg) (Model, tea.Cmd) {
	screen, ok := m.Navigator.CurrentScreen()
	if !ok {
		return m, nil
	}
	updated, cmd := screen.Update(msg)
	m.Navigator.SetCurrentScreen(updated)
	return m.dispatch(cmd)
}
