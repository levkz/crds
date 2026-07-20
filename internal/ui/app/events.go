package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/events"
	"crds/internal/ui/keymap"
	"crds/internal/ui/theme"
)

type setDecker interface {
	SetDeck(ui.DeckData)
}

type setSearchDataer interface {
	SetSearchData([]ui.CardData)
}

type setStatter interface {
	SetStats(ui.Stats)
}

type entrySetter interface {
	SetEntry(ui.CardData)
}

// dispatchEvent routes a message to its registered handler or the active screen.
func (m Model) dispatchEvent(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if screen, ok := m.Navigator.CurrentScreen(); ok {
			screen.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyMsg:
		return m.dispatchKeyEvent(msg)

	case ShowOverlayMsg:
		return m.WithOverlay(msg.Type), nil

	case HideOverlayMsg:
		return m.WithoutOverlay(), nil

	case events.ShowNotificationMsg:
		return m.WithNotification(msg.Text), nil

	case events.HideNotificationMsg:
		return m.WithoutNotification(), nil

	case SetLoadingMsg:
		return m.WithLoading(msg.Loading), nil

	case ui.NavigateToMsg:
		return m.transitionTo(msg.Screen)

	case ui.NavigateToDetailMsg:
		if detail, ok := m.Navigator.Registry().Get(ui.DetailScreen); ok {
			if setter, ok := detail.(entrySetter); ok {
				setter.SetEntry(msg.Entry)
			}
			m.Navigator.SetCurrentScreen(detail)
		}
		return m.transitionTo(msg.Screen)

	case DataLoadedMsg:
		return m.handleDataLoaded(msg)

	case DataErrorMsg:
		return m.WithNotification("Error loading "+msg.Kind.String()+": "+msg.Err.Error()), nil

	case ui.SaveAnswerMsg:
		return m, RecordAnswerCmd(m.Dispatcher, msg.CardID, msg.Grade)

	case StatsLoadedMsg:
		if screen, ok := m.Navigator.Registry().Get(m.Navigator.Current); ok {
			if setter, ok := screen.(setStatter); ok {
				setter.SetStats(msg.Stats)
			}
		}
		return m, nil

	case SavedMsg:
		return m.WithNotification("Saved " + msg.Kind.String()), nil

	case SaveErrorMsg:
		return m.WithNotification("Error saving "+msg.Kind.String()+": "+msg.Err.Error()), nil

	case ConfigUpdatedMsg:
		m.Config = msg.Config
		return m, nil

	case events.ThemeSwitchMsg:
		th, err := theme.Switch(msg.Name)
		if err != nil {
			return m.WithNotification("Theme not found: " + msg.Name), nil
		}
		ui.SetTheme(th)
		return m.WithNotification("Switched to " + msg.Name + " theme"), nil

	case events.TickMsg:
		return m, TickCmd()

	default:
		if m.Global.Overlay != NoOverlay {
			return m, nil
		}
		return m.forwardToScreen(msg)
	}
}

// handleDataLoaded stores loaded data and passes it to relevant screens.
func (m Model) handleDataLoaded(msg DataLoadedMsg) (Model, tea.Cmd) {
	switch msg.Kind {
	case MsgKindDeckList:
		names, ok := msg.Data.([]string)
		if !ok {
			return m, nil
		}
		m.AllDecks = names

		// Auto-load the first available deck
		if len(names) > 0 {
			return m, LoadDeckCmd(m.Dispatcher, names[0])
		}
		return m.WithNotification("No decks found"), nil

	case MsgKindDeck:
		deck, ok := msg.Data.(ui.DeckData)
		if !ok {
			return m, nil
		}
		m.CurrentDeck = &deck

		// Pass deck data to Quiz screen (always, so it's ready when user navigates there)
		if quiz, ok := m.Navigator.Registry().Get(ui.QuizScreen); ok {
			if setter, ok := quiz.(setDecker); ok {
				setter.SetDeck(deck)
			}
		}

		// Pass cards to Search screen
		if search, ok := m.Navigator.Registry().Get(ui.SearchScreen); ok {
			if setter, ok := search.(setSearchDataer); ok {
				setter.SetSearchData(deck.Cards)
			}
		}

		return m, nil
	}
	return m, nil
}

// dispatchKeyEvent handles global keyboard shortcuts.
func (m Model) dispatchKeyEvent(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case keymap.DefaultGlobal.Quit.Match(msg):
		return m, tea.Sequence(m.ShutdownCmd(), tea.Quit)
	case keymap.DefaultGlobal.Help.Match(msg):
		return m.WithOverlay(HelpOverlay), nil
	case keymap.DefaultGlobal.Back.Match(msg):
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
