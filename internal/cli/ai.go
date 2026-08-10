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

// printEntries renders entries as YAML to stdout.
func printEntries(entries []model.Entry) error {
	data, err := yaml.Marshal(entries)
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

// --- interpret ---

type AiInterpretCmd struct {
	Deck   string `help:"Deck for language context (optional)." completion-predictor:"deck"`
	Text   string `short:"t" help:"Free-form words (inline)."`
	File   string `short:"f" help:"Path to a text file (use - for stdin)."`
	DryRun bool   `help:"Print the prompt without calling the API."`
}

func (c *AiInterpretCmd) Run(a *app.App) error {
	raw, err := readInput(c.Text, c.File, "# Enter words, one per line.\n")
	if err != nil {
		return err
	}

	var lc ai.LanguageContext
	if c.Deck != "" {
		deck, err := loadDeckModel(a, c.Deck)
		if err != nil {
			return err
		}
		lc = ai.LanguageContext{Language: deck.Language, TranslationLanguage: deck.TranslationLanguage}
	}

	if c.DryRun {
		system, user := ai.InterpretMessages(raw, lc)
		return printPrompt(system, user)
	}

	client, err := resolveAIClient()
	if err != nil {
		return err
	}
	entries, err := ai.Interpret(context.Background(), client, raw, lc)
	if err != nil {
		return err
	}
	return printEntries(entries)
}

// --- fill ---

type AiFillCmd struct {
	Deck   string `arg:"" required:"" help:"Deck to fill for (languages + existing tags)." completion-predictor:"deck"`
	Text   string `short:"t" help:"Partial YAML entries (inline)."`
	File   string `short:"f" help:"Path to a YAML file (use - for stdin)."`
	DryRun bool   `help:"Print the prompt without calling the API."`
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

	if c.DryRun {
		system, user := ai.FillMessages(entries, ctx)
		return printPrompt(system, user)
	}

	client, err := resolveAIClient()
	if err != nil {
		return err
	}
	filled, err := ai.Fill(context.Background(), client, entries, ctx)
	if err != nil {
		return err
	}
	return printEntries(filled)
}

// --- add ---

type AiAddCmd struct {
	Deck string `arg:"" required:"" help:"Deck to append entries to." completion-predictor:"deck"`
	Text string `short:"t" help:"Words or YAML entries (inline)."`
	File string `short:"f" help:"Path to an input file (use - for stdin)."`
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
		entries, err = ai.Interpret(ctx, client, raw, lc)
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

	filled, err := ai.Fill(ctx, client, entries, fillCtx)
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