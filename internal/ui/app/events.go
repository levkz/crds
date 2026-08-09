package app

import (
	"crds/internal/stats"
	"crds/internal/ui"
	"crds/internal/ui/events"
	"crds/internal/ui/keymap"
	"crds/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
)

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
		return m.WithNotification("Error loading " + msg.Kind.String() + ": " + msg.Err.Error()), nil

	case ui.SaveAnswerMsg:
		m.AnswersRecorded = true
		return m, tea.Batch(
			RecordAnswerCmd(m.Dispatcher, msg),
			LoadDueProgressCmd(m.Dispatcher, m.State.SelectedDecks, m.State.SelectedTags),
		)

	case LoadDueProgressMsg:
		m.State.Due = msg.Due
		m.State.DeckProgress = msg.Progress
		return m, m.stateChangedCmd()

	case StatsLoadedMsg:
		m.State.Stats = &msg.Stats
		m.State.SelectionStats = msg.SelectionStats
		m.State.SelectionHistory = msg.SelectionHistory
		return m, m.stateChangedCmd()

	case SavedMsg:
		return m.WithNotification("Saved " + msg.Kind.String()), nil

	case SaveErrorMsg:
		return m.WithNotification("Error saving " + msg.Kind.String() + ": " + msg.Err.Error()), nil

	case ConfigUpdatedMsg:
		m.Config = msg.Config
		m.State.QuizMode = msg.Config.QuizMode
		return m, m.stateChangedCmd()

	case ui.StateChangedMsg:
		if m.Global.Overlay != NoOverlay {
			return m, nil
		}
		return m, m.syncActiveScreen()

	case ui.SetQuizModeMsg:
		m.State.QuizMode = msg.Mode
		return m, m.stateChangedCmd()

	case ui.RefreshStatsMsg:
		return m, FetchStatsCmd(m.Dispatcher, m.State.SelectedDecks, m.State.SelectedTags)

	case ui.RefreshWordStatsMsg:
		return m, FetchWordStatsCmd(m.Dispatcher, msg.EntryID)

	case events.ThemeSwitchMsg:
		th, err := theme.Switch(msg.Name)
		if err != nil {
			return m.WithNotification("Theme not found: " + msg.Name), nil
		}
		ui.SetTheme(th)
		cmd := func() tea.Msg {
			return SaveStateCmd(m.Dispatcher, m.State.SelectedDecks)
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

// handleDataLoaded stores loaded data in the global state snapshot and emits
// a StateChangedMsg so screens react to it.
func (m Model) handleDataLoaded(msg DataLoadedMsg) (Model, tea.Cmd) {
	switch msg.Kind {
	case MsgKindDeckList:
		names, ok := msg.Data.([]string)
		if !ok {
			return m, nil
		}
		m.State.AllDecks = names

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
		m.State.SelectedDecks = validSelected
		m.State.SelectedTags = state.SelectedTags

		// Restore theme from saved state
		if state.Theme != "" {
			if th, err := theme.Switch(state.Theme); err == nil {
				ui.SetTheme(th)
			}
		}

		// Load tags and deck-tag mapping in background
		var cmds []tea.Cmd
		cmds = append(cmds, ListAllTagsCmd(m.Dispatcher))
		cmds = append(cmds, LoadAllDeckTagsCmd(m.Dispatcher))

		// Load selected decks or show empty
		if len(validSelected) > 0 {
			cmds = append(cmds, LoadSelectedDecksCmd(m.Dispatcher, validSelected))
			cmds = append(cmds, LoadDueProgressCmd(m.Dispatcher, validSelected, m.State.SelectedTags))
		} else {
			// Handle empty selection: empty deck so screens show "No cards loaded"
			emptyDeck := ui.DeckData{}
			m.State.Deck = &emptyDeck
			m.State.DeckProgress = nil
			cmds = append(cmds, LoadDueProgressCmd(m.Dispatcher, nil, m.State.SelectedTags))
		}
		cmds = append(cmds, m.stateChangedCmd())
		return m, tea.Batch(cmds...)

	case MsgKindTags:
		tags, ok := msg.Data.([]string)
		if !ok {
			return m, nil
		}
		m.State.AllTags = tags

		// Filter saved selected tags to only available ones
		m.State.SelectedTags = filterAvailable(m.State.SelectedTags, tags)
		return m, m.stateChangedCmd()

	case MsgKindDeckTags:
		dt, ok := msg.Data.(map[string][]string)
		if !ok {
			return m, nil
		}
		m.State.AllDeckTags = dt
		return m, m.stateChangedCmd()

	case MsgKindDeck:
		deck, ok := msg.Data.(ui.DeckData)
		var progress map[string]stats.EntryProgress
		if !ok {
			dwp, ok2 := msg.Data.(DeckWithProgressMsg)
			if !ok2 {
				return m, nil
			}
			deck = dwp.Deck
			progress = dwp.Progress
		}
		m.State.Deck = &deck
		m.State.DeckProgress = progress
		return m, m.stateChangedCmd()
	}
	return m, nil
}

func (m Model) handleDeckSelectionWithTags(selected []string, selectedTags []string) (Model, tea.Cmd) {
	m.State.SelectedDecks = selected
	m.State.SelectedTags = selectedTags
	var cmds []tea.Cmd
	cmds = append(cmds, func() tea.Msg {
		return SaveStateCmd(m.Dispatcher, selected, selectedTags...)
	})
	cmds = append(cmds, ResetSessionCmd(m.Dispatcher))
	if len(selected) > 0 {
		cmds = append(cmds, LoadSelectedDecksCmd(m.Dispatcher, selected))
		cmds = append(cmds, LoadDueProgressCmd(m.Dispatcher, selected, selectedTags))
	} else {
		m.State.Deck = nil
		cmds = append(cmds, func() tea.Msg {
			return DataLoadedMsg{Kind: MsgKindDeck, Data: ui.DeckData{}}
		})
		cmds = append(cmds, LoadDueProgressCmd(m.Dispatcher, nil, selectedTags))
	}
	cmds = append(cmds, m.stateChangedCmd())
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

// stateChangedCmd returns a command that notifies screens of the current
// AppState snapshot so they can recompute derived state. It captures the
// snapshot as of the moment the mutation happened.
func (m Model) stateChangedCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.StateChangedMsg{State: m.State}
	}
}

// syncActiveScreen pushes the current AppState snapshot to the active screen
// if it implements StateSyncer.
func (m Model) syncActiveScreen() tea.Cmd {
	screen, ok := m.Navigator.CurrentScreen()
	if !ok {
		return nil
	}
	if s, ok := screen.(ui.StateSyncer); ok {
		return s.SyncState(m.State)
	}
	return nil
}
