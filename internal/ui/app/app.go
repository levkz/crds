package app

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/ui"
	"crds/internal/ui/screens"
	"crds/internal/ui/theme"
	nav "crds/internal/ui/navigation"
)

// New creates and initializes the root UI application model with injected dependencies
// and an optional config override. Pass DefaultConfig() to use defaults.
func New(deps Dependencies, cfg Config) Model {
	n := nav.New(ui.HomeScreen)
	reg := nav.NewRegistry()
	reg.Register(ui.HomeScreen, screens.NewHome())
	reg.Register(ui.QuizScreen, screens.NewQuiz())
	reg.Register(ui.SearchScreen, screens.NewSearch())
	reg.Register(ui.StatisticsScreen, screens.NewStatistics())
	reg.Register(ui.SettingsScreen, screens.NewSettings())
	reg.Register(ui.DetailScreen, screens.NewDetail())
	n.SetRegistry(reg)

	if cfg.ThemePath != "" {
		t, err := theme.LoadTheme(cfg.ThemePath)
		if err != nil {
			log.Printf("warning: failed to load theme %q: %v", cfg.ThemePath, err)
		} else {
			theme.Register("_loaded", t)
			if _, err := theme.Switch("_loaded"); err == nil {
				ui.SetTheme(t)
			}
		}
	}

	return Model{
		Config:     cfg,
		Navigator:  n,
		Dispatcher: &Dispatcher{Decks: deps.Decks, Progress: deps.Progress},
	}
}

func (m Model) dispatch(cmd tea.Cmd) (Model, tea.Cmd) {
	if cmd == nil {
		return m, nil
	}
	return m, cmd
}

// Run runs the Bubble Tea program with injected dependencies and config.
func Run(deps Dependencies, cfg Config) error {
	p := tea.NewProgram(
		New(deps, cfg),
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}

// RunWithDefaults runs the Bubble Tea program with default config.
func RunWithDefaults(deps Dependencies) error {
	return Run(deps, DefaultConfig())
}
