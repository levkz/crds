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

type setDeckSelectionDataer interface {
	SetData(deckItems, selectedDecks, tagItems, selectedTags []string, deckTags map[string][]string)
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
		if msg.Screen == ui.DecksScreen {
			return m.pushTo(msg.Screen)
		}
		if m.AnswersRecorded && m.isQuizScreen() {
			m.PendingTarget = &msg.Screen
			return m.WithOverlay(ConfirmOverlay), nil
		}
		return m.transitionTo(msg.Screen)

	case ConfirmYesMsg:
		if m.PendingDeckSelection != nil {
			m, cmd := m.handleDeckSelectionWithTags(m.PendingDeckSelection, m.PendingTagSelection)
			m.PendingDeckSelection = nil
			m.PendingTagSelection = nil
			return m.WithoutOverlay(), cmd
		}
		if m.PendingTarget == nil {
			return m.WithoutOverlay(), nil
		}
		target := *m.PendingTarget
		m.PendingTarget = nil
		m.AnswersRecorded = false
		return m.WithoutOverlay().transitionTo(target)

	case ConfirmNoMsg:
		m.PendingDeckSelection = nil
		m.PendingTagSelection = nil
		m.PendingTarget = nil
		return m.WithoutOverlay(), nil

	case ui.NavigateToDetailMsg:
		if detail, ok := m.Navigator.Registry().Get(ui.DetailScreen); ok {
			if setter, ok := detail.(entrySetter); ok {
				setter.SetEntry(msg.Entry)
			}
		}
		return m.pushTo(msg.Screen)

	case ui.DeckSelectionChangedMsg:
		if m.AnswersRecorded && m.isQuizScreen() && m.isQuizInProgress() {
			m.PendingDeckSelection = msg.Selected
			m.PendingTagSelection = msg.SelectedTags
			return m.WithOverlay(ConfirmOverlay), nil
		}
		m, cmd := m.handleDeckSelectionWithTags(msg.Selected, msg.SelectedTags)
		if m.Navigator.CanGoBack() {
			navM, navCmd := m.popToPrevious()
			return navM, tea.Batch(cmd, navCmd)
		}
		return m, tea.Batch(cmd, func() tea.Msg {
			return ui.NavigateToMsg{Screen: ui.HomeScreen}
		})

	case DataLoadedMsg:
		return m.handleDataLoaded(msg)

	case DataErrorMsg:
		return m.WithNotification("Error loading "+msg.Kind.String()+": "+msg.Err.Error()), nil

	case ui.SaveAnswerMsg:
		m.AnswersRecorded = true
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

		// Pass deck list + selection + tags to DeckSelect screen
		if ds, ok := m.Navigator.Registry().Get(ui.DecksScreen); ok {
			if setter, ok := ds.(setDeckSelectionDataer); ok {
				m.SelectedTags = state.SelectedTags
				setter.SetData(names, validSelected, m.AllTags, m.SelectedTags, m.AllDeckTags)
			}
		}

		// Also pass to old Decks screen for compatibility during transition
		if decksScreen, ok := m.Navigator.Registry().Get(ui.DecksScreen); ok {
			if setter, ok := decksScreen.(setDecksLister); ok {
				setter.SetDecks(names, validSelected)
			}
		}

		// Load tags and deck-tag mapping in background
		var cmds []tea.Cmd
		cmds = append(cmds, ListAllTagsCmd(m.Dispatcher))
		cmds = append(cmds, LoadAllDeckTagsCmd(m.Dispatcher))

		// Load selected decks or show empty
		if len(validSelected) > 0 {
			cmds = append(cmds, LoadSelectedDecksCmd(m.Dispatcher, validSelected))
		} else {
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
		}
		return m, tea.Batch(cmds...)

	case MsgKindTags:
		tags, ok := msg.Data.([]string)
		if !ok {
			return m, nil
		}
		m.AllTags = tags

		// Filter saved selected tags to only available ones
		m.SelectedTags = filterAvailable(m.SelectedTags, tags)

		// Pass tags to DeckSelect screen
		if ds, ok := m.Navigator.Registry().Get(ui.DecksScreen); ok {
			if setter, ok := ds.(setDeckSelectionDataer); ok {
				setter.SetData(m.AllDecks, m.SelectedDecks, tags, m.SelectedTags, m.AllDeckTags)
			}
		}
		return m, nil

	case MsgKindDeckTags:
		dt, ok := msg.Data.(map[string][]string)
		if !ok {
			return m, nil
		}
		m.AllDeckTags = dt

		// Pass deck tags to DeckSelect screen
		if ds, ok := m.Navigator.Registry().Get(ui.DecksScreen); ok {
			if setter, ok := ds.(setDeckSelectionDataer); ok {
				setter.SetData(m.AllDecks, m.SelectedDecks, m.AllTags, m.SelectedTags, dt)
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

func (m Model) handleDeckSelection(selected []string) (Model, tea.Cmd) {
	m.SelectedDecks = selected
	var cmds []tea.Cmd
	cmds = append(cmds, func() tea.Msg {
		return SaveStateCmd(m.Dispatcher, selected)
	})
	cmds = append(cmds, ResetSessionCmd(m.Dispatcher))
	if len(selected) > 0 {
		cmds = append(cmds, LoadSelectedDecksCmd(m.Dispatcher, selected))
	} else {
		m.CurrentDeck = nil
		cmds = append(cmds, func() tea.Msg {
			return DataLoadedMsg{Kind: MsgKindDeck, Data: ui.DeckData{}}
		})
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleDeckSelectionWithTags(selected []string, selectedTags []string) (Model, tea.Cmd) {
	m.SelectedDecks = selected
	m.SelectedTags = selectedTags
	var cmds []tea.Cmd
	cmds = append(cmds, func() tea.Msg {
		return SaveStateCmd(m.Dispatcher, selected, selectedTags...)
	})
	cmds = append(cmds, ResetSessionCmd(m.Dispatcher))
	if len(selected) > 0 {
		cmds = append(cmds, LoadSelectedDecksCmd(m.Dispatcher, selected))
	} else {
		m.CurrentDeck = nil
		cmds = append(cmds, func() tea.Msg {
			return DataLoadedMsg{Kind: MsgKindDeck, Data: ui.DeckData{}}
		})
	}
	return m, tea.Batch(cmds...)
}

// dispatchKeyEvent handles global keyboard shortcuts.
func (m Model) dispatchKeyEvent(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case keymap.DefaultGlobal.Quit.Match(msg):
		return m, tea.Sequence(m.ShutdownCmd(), tea.Quit)
	case keymap.DefaultGlobal.Help.Match(msg):
		return m.WithOverlay(HelpOverlay), nil
	case keymap.DefaultGlobal.DeckSelect.Match(msg):
		return m.pushTo(ui.DecksScreen)

	case keymap.DefaultGlobal.Back.Match(msg):
		if m.Global.Overlay == ConfirmOverlay {
			return m, func() tea.Msg { return ConfirmNoMsg{} }
		}
		if m.Global.Overlay != NoOverlay {
			return m.WithoutOverlay(), nil
		}
		if screen, ok := m.Navigator.CurrentScreen(); ok {
			if handler, ok := screen.(ui.BackHandler); ok && handler.HandleBack() {
				m.Navigator.SetCurrentScreen(screen)
				return m, nil
			}
		}
		if m.Navigator.CanGoBack() {
			return m.popToPrevious()
		}
		return m.transitionTo(ui.HomeScreen)
	}

	// Route keys to confirm overlay if active
	if m.Global.Overlay == ConfirmOverlay {
		switch {
		case keymap.DefaultList.Select.Match(msg):
			return m, func() tea.Msg { return ConfirmYesMsg{} }
		case msg.String() == "y" || msg.String() == "Y":
			return m, func() tea.Msg { return ConfirmYesMsg{} }
		case msg.String() == "n" || msg.String() == "N":
			return m, func() tea.Msg { return ConfirmNoMsg{} }
		}
		return m, nil
	}

	return m.forwardToScreen(msg)
}

func filterAvailable(items, available []string) []string {
	avail := make(map[string]bool, len(available))
	for _, a := range available {
		avail[a] = true
	}
	var filtered []string
	for _, s := range items {
		if avail[s] {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (m Model) isQuizScreen() bool {
	return m.Navigator.Current == ui.QuizScreen || m.Navigator.Current == ui.TypingQuizScreen
}

func (m Model) isQuizInProgress() bool {
	screen, ok := m.Navigator.CurrentScreen()
	if !ok {
		return false
	}
	if q, ok := screen.(ui.QuizInProgressChecker); ok {
		return q.IsInProgress()
	}
	return false
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
