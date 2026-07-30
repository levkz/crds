package app

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"crds/internal/config"
	"crds/internal/ui"
	"crds/internal/ui/keymap"
	"crds/internal/ui/screens"
	"crds/internal/ui/theme"
	nav "crds/internal/ui/navigation"
)

// New creates and initializes the root UI application model with injected dependencies
// and an optional config override. Pass DefaultConfig() to use defaults.
func New(deps Dependencies, cfg Config) Model {
	if err := config.EnsureDefaultFiles(); err != nil {
		log.Printf("warning: config init: %v", err)
	}

	// Load and apply keymap overrides from ~/.config/crds/keymaps.yaml
	keyPath, err := config.KeymapsPath()
	if err == nil {
		kmCfg, err := config.LoadKeymapConfig(keyPath)
		if err != nil {
			log.Printf("warning: loading keymaps: %v", err)
		} else if kmCfg != nil {
			keymap.ApplyDefaultOverrides(*kmCfg)
		}
	}

	// Load user themes from ~/.config/crds/themes/
	if err := config.LoadUserThemes(); err != nil {
		log.Printf("warning: loading user themes: %v", err)
	}

	// Load and apply app config from ~/.config/crds/config.yaml
	cfgPath, err := config.ConfigPath()
	var y *config.ConfigYAML
	if err == nil {
		y, err = config.LoadConfigYAML(cfgPath)
		if err != nil {
			log.Printf("warning: loading config: %v", err)
		} else {
			cfg = cfg.ApplyYAML(y)
		}
	}

	n := nav.New(ui.HomeScreen)
	reg := nav.NewRegistry()
	reg.Register(ui.HomeScreen, screens.NewHome())
	reg.Register(ui.QuizScreen, screens.NewQuiz())
	reg.Register(ui.DecksScreen, screens.NewDeckSelect())
	reg.Register(ui.TypingQuizScreen, screens.NewTypingQuiz())
	reg.Register(ui.SearchScreen, screens.NewSearch())
	reg.Register(ui.StatisticsScreen, screens.NewStatistics())
	reg.Register(ui.SettingsScreen, screens.NewSettings())
	reg.Register(ui.DetailScreen, screens.NewDetail())
	reg.Register(ui.PaletteScreen, screens.NewPaletteModel())
	n.SetRegistry(reg)

	// Apply theme from config, preferring explicit ThemePath over config.yaml's theme name
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
	} else if y != nil && y.Theme != "" {
		if _, err := theme.Switch(y.Theme); err != nil {
			log.Printf("warning: theme %q not found, using default", y.Theme)
		}
	}

	// Wire session and typing support if Store provides them
	var sessions SessionManager
	if sm, ok := deps.Progress.(SessionManager); ok {
		sessions = sm
	}
	var typing TypingRecorder
	if tr, ok := deps.Progress.(TypingRecorder); ok {
		typing = tr
	}
	var tags TagProvider
	if tp, ok := deps.Decks.(TagProvider); ok {
		tags = tp
	} else if deps.Tags != nil {
		tags = deps.Tags
	}

	return Model{
		Config:      cfg,
		Navigator:   n,
		Dispatcher: &Dispatcher{
			Decks:    deps.Decks,
			Progress: deps.Progress,
			Stats:    deps.Stats,
			State:    deps.State,
			Sessions: sessions,
			Typing:   typing,
			Tags:     tags,
		},
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
