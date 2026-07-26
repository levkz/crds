package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"crds/internal/app"
	"crds/internal/editor"
	"crds/internal/model"
	"crds/internal/parser"
)

type TermCmd struct {
	Add  TermAddCmd  `cmd:"" help:"Add a new term."`
	Rm   TermRmCmd   `cmd:"" help:"Remove a term."`
	Edit TermEditCmd `cmd:"" help:"Edit a term."`
}

type TermAddCmd struct {
	Deck         string   `arg:"" required:"" help:"Deck to add the term to." completion-predictor:"deck"`
	TermText     string   `short:"t" help:"Term text (inline)."`
	File         string   `short:"f" help:"Path to YAML file (use - for stdin)."`
	Translations string   `help:"Comma-separated translations."`
	Examples     string   `help:"Comma-separated example sentences."`
	Tags         []string `help:"Tags to apply (repeatable)."`
}

func (c *TermAddCmd) Run(a *app.App) error {
	var entry model.Entry

	switch {
	case c.File != "":
		var data []byte
		var err error
		if c.File == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(c.File)
		}
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		if err := yaml.Unmarshal(data, &entry); err != nil {
			return fmt.Errorf("parse entry: %w", err)
		}

	case c.TermText != "":
		entry.Term = c.TermText
		if c.Translations != "" {
			for _, t := range strings.Split(c.Translations, ",") {
				entry.Translations = append(entry.Translations, model.Translation{Text: strings.TrimSpace(t)})
			}
		}
		if c.Examples != "" {
			for _, ex := range strings.Split(c.Examples, ",") {
				entry.Examples = append(entry.Examples, model.Example{Text: strings.TrimSpace(ex)})
			}
		}
		entry.Tags = c.Tags

	default:
		content, err := editor.Edit(editor.EntryTemplate())
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal([]byte(content), &entry); err != nil {
			return fmt.Errorf("parse entry: %w", err)
		}
	}

	return a.Store.AddEntry(c.Deck, entry, a.DataDir)
}

type TermRmCmd struct {
	Deck   string `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID string `arg:"" required:"" help:"Term ID to remove." completion-predictor:"term"`
	Force  bool   `short:"f" help:"Skip confirmation."`
}

func (c *TermRmCmd) Run(a *app.App) error {
	if !c.Force {
		fmt.Printf("Remove term %q from deck %q? [y/N] ", c.TermID, c.Deck)
		var answer string
		if _, err := fmt.Scan(&answer); err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	return a.Store.RemoveEntry(c.Deck, c.TermID, a.DataDir)
}

type TermEditCmd struct {
	Deck     string `arg:"" required:"" help:"Deck containing the term." completion-predictor:"deck"`
	TermID   string `arg:"" required:"" help:"Term ID to edit." completion-predictor:"term"`
	TermText string `short:"t" help:"New term text (inline)."`
	File     string `short:"f" help:"Path to updated YAML file (use - for stdin)."`
}

func (c *TermEditCmd) Run(a *app.App) error {
	path := filepath.Join(a.DataDir, c.Deck+".yaml")
	deck, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse deck %q: %w", c.Deck, err)
	}

	var entry *model.Entry
	for i := range deck.Entries {
		if deck.Entries[i].ID == c.TermID {
			entry = &deck.Entries[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("entry %q not found in deck %q", c.TermID, c.Deck)
	}

	var modified *model.Entry

	switch {
	case c.File != "":
		var data []byte
		var err error
		if c.File == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(c.File)
		}
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		var parsed model.Entry
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("parse entry: %w", err)
		}
		parsed.ID = c.TermID
		modified = &parsed

	case c.TermText != "":
		entry.Term = c.TermText
		return a.Store.UpdateEntry(c.Deck, c.TermID, *entry, a.DataDir)

	default:
		modified, err = editor.EditEntry(entry)
		if err != nil {
			return err
		}
	}

	return a.Store.UpdateEntry(c.Deck, c.TermID, *modified, a.DataDir)
}
