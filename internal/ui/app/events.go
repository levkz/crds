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

type setDecksLister interface {
	SetDecks(decks []string, selected []string)
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

	case ui.DeckSelectionChangedMsg:
		m.SelectedDecks = msg.Selected
		var cmds []tea.Cmd
		cmds = append(cmds, func() tea.Msg {
			return SaveStateCmd(m.Dispatcher, msg.Selected)
		})
		cmds = append(cmds, ResetSessionCmd(m.Dispatcher))
		if len(msg.Selected) > 0 {
			cmds = append(cmds, LoadSelectedDecksCmd(m.Dispatcher, msg.Selected))
		} else {
			m.CurrentDeck = nil
			cmds = append(cmds, func() tea.Msg {
				return DataLoadedMsg{Kind: MsgKindDeck, Data: ui.DeckData{}}
			})
		}
		return m, tea.Batch(cmds...)

	case DataLoadedMsg:
		return m.handleDataLoaded(msg)

	case DataErrorMsg:
		return m.WithNotification("Error loading "+msg.Kind.String()+": "+msg.Err.Error()), nil

	case ui.SaveAnswerMsg:
		return m, RecordAnswerCmd(m.Dispatcher, msg)

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
		cmd := func() tea.Msg {
			return SaveStateCmd(m.Dispatcher, m.SelectedDecks)
		}
		return m.WithNotification("Switched to " + msg.Name + " theme"), cmd

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

		// Load saved state
		state, err := m.Dispatcher.State.Load()
		if err != nil {
			return m.WithNotification("Error loading state: " + err.Error()), nil
		}

		// Filter selected decks to only available ones
		available := make(map[string]bool, len(names))
		for _, n := range names {
			available[n] = true
		}
		var validSelected []string
		for _, s := range state.SelectedDecks {
			if available[s] {
				validSelected = append(validSelected, s)
			}
		}
		m.SelectedDecks = validSelected

		// Restore theme from saved state
		if state.Theme != "" {
			if th, err := theme.Switch(state.Theme); err == nil {
				ui.SetTheme(th)
			}
		}

		// Pass deck list + selection to Decks screen
		if decksScreen, ok := m.Navigator.Registry().Get(ui.DecksScreen); ok {
			if setter, ok := decksScreen.(setDecksLister); ok {
				setter.SetDecks(names, validSelected)
			}
		}

		// Load selected decks or show empty
		if len(validSelected) > 0 {
			return m, LoadSelectedDecksCmd(m.Dispatcher, validSelected)
		}

		// Handle empty selection: pass empty deck to screens so they show "No cards loaded"
		emptyDeck := ui.DeckData{}
		m.CurrentDeck = &emptyDeck
		if quiz, ok := m.Navigator.Registry().Get(ui.QuizScreen); ok {
			if setter, ok := quiz.(setDecker); ok {
				setter.SetDeck(emptyDeck)
			}
		}
		if typingQuiz, ok := m.Navigator.Registry().Get(ui.TypingQuizScreen); ok {
			if setter, ok := typingQuiz.(setDecker); ok {
				setter.SetDeck(emptyDeck)
			}
		}
		return m, nil

	case MsgKindDeck:
		deck, ok := msg.Data.(ui.DeckData)
		if !ok {
			return m, nil
		}
		m.CurrentDeck = &deck

		// Pass deck data to Quiz screens (always, so they're ready when user navigates there)
		if quiz, ok := m.Navigator.Registry().Get(ui.QuizScreen); ok {
			if setter, ok := quiz.(setDecker); ok {
				setter.SetDeck(deck)
			}
		}
		if typingQuiz, ok := m.Navigator.Registry().Get(ui.TypingQuizScreen); ok {
			if setter, ok := typingQuiz.(setDecker); ok {
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
