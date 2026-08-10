package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"crds/internal/ai"
	"crds/internal/app"
	"crds/internal/config"
	"crds/internal/editor"
	"crds/internal/model"
	"crds/internal/parser"
)

type AiCmd struct {
	Interpret AiInterpretCmd `cmd:"" help:"Convert free-form text into YAML entries."`
	Fill      AiFillCmd      `cmd:"" help:"Complete partial YAML entries with the AI agent."`
	Add       AiAddCmd       `cmd:"" help:"Interpret words (or YAML), fill them in, and append to a deck."`
}

// resolveAIClient reads the `ai:` config block, resolves defaults/overrides,
// and returns a ready client. It is a variable so tests can substitute a fake
// and never touch the network or the user's real config.
var resolveAIClient = func() (ai.Client, error) {
	fileCfg, err := aiConfigFromFile()
	if err != nil {
		return nil, err
	}
	resolved, err := ai.Resolve(fileCfg)
	if err != nil {
		return nil, err
	}
	return ai.NewClient(resolved), nil
}

// aiConfigFromFile reads the `ai:` block from the user config (if present).
func aiConfigFromFile() (ai.Config, error) {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return ai.Config{}, err
	}
	cfg, err := config.LoadConfigYAML(cfgPath)
	if err != nil {
		return ai.Config{}, err
	}
	if cfg == nil || cfg.AI == nil {
		return ai.Config{}, nil
	}
	return ai.Config{
		Provider: cfg.AI.Provider,
		Model:    cfg.AI.Model,
		APIKey:   cfg.AI.APIKey,
		BaseURL:  cfg.AI.BaseURL,
	}, nil
}

// loadDeckModel reads a deck's YAML directly (ui.DeckData drops the languages).
func loadDeckModel(a *app.App, deckID string) (*model.Deck, error) {
	path := filepath.Join(a.DataDir, deckID+".yaml")
	deck, err := parser.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("load deck %q: %w", deckID, err)
	}
	return deck, nil
}

// readInput resolves --text / --file / editor into raw input text.
func readInput(text, file string, editorTemplate string) (string, error) {
	switch {
	case text != "":
		return text, nil
	case file != "":
		var data []byte
		var err error
		if file == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(file)
		}
		if err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		return string(data), nil
	default:
		return editor.Edit(editorTemplate)
	}
}

// entryView renders only populated fields so output is clean regardless of
// what the model returned. model.Entry lacks omitempty tags.
type entryView struct {
	ID           string              `yaml:"id,omitempty"`
	Term         string              `yaml:"term"`
	Translations []model.Translation `yaml:"translations"`
	Examples     []model.Example     `yaml:"examples,omitempty"`
	Tags         []string            `yaml:"tags,omitempty"`
	Notes        string              `yaml:"notes,omitempty"`
}

// printEntries renders entries as YAML to stdout, omitting empty fields.
func printEntries(entries []model.Entry) error {
	views := make([]entryView, len(entries))
	for i, e := range entries {
		views[i] = entryView{
			ID:           e.ID,
			Term:         e.Term,
			Translations: e.Translations,
			Examples:     e.Examples,
			Tags:         e.Tags,
			Notes:        e.Notes,
		}
	}
	data, err := yaml.Marshal(views)
	if err != nil {
		return fmt.Errorf("render entries: %w", err)
	}
	fmt.Println(strings.TrimSuffix(string(data), "\n"))
	return nil
}

// deckContextForFill gathers everything the filler needs from a deck.
func deckContextForFill(a *app.App, deckID string) (ai.DeckContext, error) {
	deck, err := loadDeckModel(a, deckID)
	if err != nil {
		return ai.DeckContext{}, err
	}

	tags, err := a.Store.ListDeckTags(deckID)
	if err != nil {
		return ai.DeckContext{}, fmt.Errorf("list deck tags: %w", err)
	}

	return ai.DeckContext{
		Language:            deck.Language,
		TranslationLanguage: deck.TranslationLanguage,
		AllowedTags:         tags,
		Samples:             pickSamples(deck.Entries, 3),
	}, nil
}

// pickSamples returns up to n existing entries suited to show conventions:
// entries with examples and tags first, filled with any remaining entries.
func pickSamples(entries []model.Entry, n int) []model.Entry {
	var withData, rest []model.Entry
	for _, e := range entries {
		if len(e.Examples) > 0 && len(e.Tags) > 0 {
			withData = append(withData, e)
		} else {
			rest = append(rest, e)
		}
	}
	combined := append(withData, rest...)
	if len(combined) > n {
		combined = combined[:n]
	}
	return combined
}

// applyLanguageOverrides lets --translate-from/--translate-to take precedence
// over the deck's language pair.
func applyLanguageOverrides(from, to string, dc ai.DeckContext) ai.DeckContext {
	if from != "" {
		dc.Language = from
	}
	if to != "" {
		dc.TranslationLanguage = to
	}
	return dc
}

// --- interpret ---

type AiInterpretCmd struct {
	Deck          string `help:"Deck for language context (optional)." completion-predictor:"deck"`
	Text          string `short:"t" help:"Free-form words (inline)."`
	File          string `short:"f" help:"Path to a text file (use - for stdin)."`
	TranslateFrom string `short:"F" help:"Source language for terms/examples (overrides the deck)."`
	TranslateTo   string `short:"T" help:"Target language for translations (overrides the deck)."`
	DryRun        bool   `help:"Print the prompt without calling the API."`
	Minimal       bool   `help:"Bare term + translations only (the default)."`
	Full          bool   `help:"Full entries: at least 4 example uses, notes, and tags from the deck allowlist."`
	Msg           string `help:"Extra instruction passed to the model."`
}

func (c *AiInterpretCmd) Run(a *app.App) error {
	if c.Minimal && c.Full {
		return fmt.Errorf("--minimal and --full are mutually exclusive")
	}

	raw, err := readInput(c.Text, c.File, "# Enter words, one per line.\n")
	if err != nil {
		return err
	}

	if c.Full {
		var dc ai.DeckContext
		if c.Deck != "" {
			dc, err = deckContextForFill(a, c.Deck)
			if err != nil {
				return err
			}
		}
		dc = applyLanguageOverrides(c.TranslateFrom, c.TranslateTo, dc)

		if c.DryRun {
			system, user := ai.InterpretFullMessages(raw, dc, c.Msg)
			return printPrompt(system, user)
		}

		client, err := resolveAIClient()
		if err != nil {
			return err
		}
		entries, err := ai.InterpretFull(context.Background(), client, raw, dc, c.Msg)
		if err != nil {
			return err
		}
		return printEntries(entries)
	}

	var lc ai.LanguageContext
	if c.Deck != "" {
		deck, err := loadDeckModel(a, c.Deck)
		if err != nil {
			return err
		}
		lc = ai.LanguageContext{Language: deck.Language, TranslationLanguage: deck.TranslationLanguage}
	}
	if c.TranslateFrom != "" {
		lc.Language = c.TranslateFrom
	}
	if c.TranslateTo != "" {
		lc.TranslationLanguage = c.TranslateTo
	}

	if c.DryRun {
		system, user := ai.InterpretMessages(raw, lc, c.Msg)
		return printPrompt(system, user)
	}

	client, err := resolveAIClient()
	if err != nil {
		return err
	}
	entries, err := ai.Interpret(context.Background(), client, raw, lc, c.Msg)
	if err != nil {
		return err
	}
	return printEntries(entries)
}

// --- fill ---

type AiFillCmd struct {
	Deck          string `arg:"" required:"" help:"Deck to fill for (languages + existing tags)." completion-predictor:"deck"`
	Text          string `short:"t" help:"Partial YAML entries (inline)."`
	File          string `short:"f" help:"Path to a YAML file (use - for stdin)."`
	TranslateFrom string `short:"F" help:"Source language for terms/examples (overrides the deck)."`
	TranslateTo   string `short:"T" help:"Target language for translations (overrides the deck)."`
	DryRun        bool   `help:"Print the prompt without calling the API."`
	Msg           string `help:"Extra instruction passed to the model."`
}

func (c *AiFillCmd) Run(a *app.App) error {
	raw, err := readInput(c.Text, c.File, editor.EntryTemplate())
	if err != nil {
		return err
	}
	entries, err := ai.ParseEntries(raw)
	if err != nil {
		return fmt.Errorf("parse input: %w", err)
	}

	ctx, err := deckContextForFill(a, c.Deck)
	if err != nil {
		return err
	}
	ctx = applyLanguageOverrides(c.TranslateFrom, c.TranslateTo, ctx)

	if c.DryRun {
		system, user := ai.FillMessages(entries, ctx, c.Msg)
		return printPrompt(system, user)
	}

	client, err := resolveAIClient()
	if err != nil {
		return err
	}
	filled, err := ai.Fill(context.Background(), client, entries, ctx, c.Msg)
	if err != nil {
		return err
	}
	return printEntries(filled)
}

// --- add ---

type AiAddCmd struct {
	Deck          string `arg:"" required:"" help:"Deck to append entries to." completion-predictor:"deck"`
	Text          string `short:"t" help:"Words or YAML entries (inline)."`
	File          string `short:"f" help:"Path to an input file (use - for stdin)."`
	TranslateFrom string `short:"F" help:"Source language for terms/examples (overrides the deck)."`
	TranslateTo   string `short:"T" help:"Target language for translations (overrides the deck)."`
	Msg           string `help:"Extra instruction passed to the model."`
}

func (c *AiAddCmd) Run(a *app.App) error {
	raw, err := readInput(c.Text, c.File, "# Enter words, one per line.\n")
	if err != nil {
		return err
	}

	deck, err := loadDeckModel(a, c.Deck)
	if err != nil {
		return err
	}

	client, err := resolveAIClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	var entries []model.Entry
	if ai.IsStructuredInput(raw) {
		entries, err = ai.ParseEntries(raw)
		if err != nil {
			return fmt.Errorf("parse input: %w", err)
		}
	} else {
		lc := ai.LanguageContext{Language: deck.Language, TranslationLanguage: deck.TranslationLanguage}
		entries, err = ai.Interpret(ctx, client, raw, lc, c.Msg)
		if err != nil {
			return err
		}
	}

	tags, err := a.Store.ListDeckTags(c.Deck)
	if err != nil {
		return fmt.Errorf("list deck tags: %w", err)
	}
	fillCtx := ai.DeckContext{
		Language:            deck.Language,
		TranslationLanguage: deck.TranslationLanguage,
		AllowedTags:         tags,
		Samples:             pickSamples(deck.Entries, 3),
	}
	fillCtx = applyLanguageOverrides(c.TranslateFrom, c.TranslateTo, fillCtx)

	filled, err := ai.Fill(ctx, client, entries, fillCtx, c.Msg)
	if err != nil {
		return err
	}

	return reviewAndAppend(a, c.Deck, filled)
}

// reviewAndAppend shows the proposed entries and lets the user append, edit
// them again in $EDITOR, or discard.
func reviewAndAppend(a *app.App, deckID string, entries []model.Entry) error {
	for {
		if err := printEntries(entries); err != nil {
			return err
		}

		fmt.Print("[a]ppend / [e]dit / [d]iscard: ")
		var answer string
		if _, err := fmt.Scanln(&answer); err != nil {
			return fmt.Errorf("read choice: %w", err)
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "a":
			if err := a.Store.AppendEntries(deckID, entries, a.DataDir); err != nil {
				return err
			}
			fmt.Printf("Appended %d entry/entries to deck %q.\n", len(entries), deckID)
			return nil
		case "e":
			data, err := yaml.Marshal(entries)
			if err != nil {
				return fmt.Errorf("render entries: %w", err)
			}
			edited, err := editor.Edit(string(data))
			if err != nil {
				return err
			}
			entries, err = ai.ParseEntries(edited)
			if err != nil {
				return fmt.Errorf("parse edited entries: %w", err)
			}
		case "d":
			fmt.Println("Discarded.")
			return nil
		default:
			fmt.Println("Unknown choice; enter a, e, or d.")
		}
	}
}

func printPrompt(system, user string) error {
	fmt.Println("=== SYSTEM ===")
	fmt.Println(system)
	fmt.Println("=== USER ===")
	fmt.Println(user)
	return nil
}