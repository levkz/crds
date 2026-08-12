package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"crds/internal/ai"
	"crds/internal/app"
	"crds/internal/storage"
)

// resolveDeck guesses and interactively resolves the deck to append/fill for
// when none was given. It asks the model to match the raw input against the
// existing deck list, confirms a match with the user, and otherwise lets the
// user create a new deck (proposed or typed name) or pick an existing one.
// Returns errAborted when the user cancels.
func resolveDeck(a *app.App, client ai.Client, raw string, msg, from, to string) (string, error) {
	decks, err := a.Store.ListDecksWithStats()
	if err != nil {
		return "", fmt.Errorf("list decks: %w", err)
	}

	infos := make([]ai.DeckInfo, len(decks))
	for i, d := range decks {
		infos[i] = ai.DeckInfo{
			ID:                  d.ID,
			Name:                d.Name,
			Language:            d.Language,
			TranslationLanguage: d.TranslationLanguage,
		}
	}

	res, err := ai.SuggestDeck(context.Background(), client, infos, raw, msg)
	if err != nil {
		return "", fmt.Errorf("suggest deck: %w", err)
	}

	if res.Deck != "" {
		if ok, err := promptYesNo(fmt.Sprintf("Did you mean deck %q%s?", res.Deck, describeDeck(res.Deck, decks))); err != nil {
			return "", err
		} else if ok {
			return res.Deck, nil
		}
	}

	return resolveDeckMenu(a, res.Proposed, decks, from, to)
}

// describeDeck formats "(fr -> en, 24 entries)" for a deck in the list, or an
// empty string when the deck is not listed.
func describeDeck(id string, decks []storage.DeckSummary) string {
	for _, d := range decks {
		if d.ID == id {
			return fmt.Sprintf(" (%s \u2192 %s, %d entries)", d.Language, d.TranslationLanguage, d.EntryCount)
		}
	}
	return ""
}

// resolveDeckMenu presents the no-match options: create a deck with the
// proposed name, type a new name, select an existing deck, or abort.
func resolveDeckMenu(a *app.App, proposal *ai.DeckProposal, decks []storage.DeckSummary, from, to string) (string, error) {
	for {
		fmt.Println("No existing deck clearly matches this input.")
		if proposal != nil {
			fmt.Printf("[c] Create deck %q (%s \u2192 %s)\n", proposal.Name, proposal.Language, proposal.TranslationLanguage)
		} else {
			fmt.Println("[c] Create a deck with a proposed name")
		}
		fmt.Println("[n] Enter a new deck name manually")
		fmt.Println("[s] Select an existing deck (tab to autocomplete)")
		fmt.Println("[a] Abort")

		choice, err := promptReadLine("Choice [c/n/s/a]: ", nil)
		if err != nil {
			return "", err
		}
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "c":
			created, err := createFromProposal(a, proposal, from, to)
			if err != nil {
				return "", err
			}
			if created != "" {
				return created, nil
			}
		case "n":
			created, err := createFromName(a, proposal, from, to)
			if err != nil {
				return "", err
			}
			if created != "" {
				return created, nil
			}
		case "s":
			deckID, err := selectDeckID(a, decks)
			if err != nil {
				return "", err
			}
			if deckID != "" {
				return deckID, nil
			}
		case "a", "q":
			fmt.Println("Aborted.")
			return "", errAborted
		default:
			fmt.Println("Unknown choice; enter c, n, s, or a.")
		}
	}
}

// createFromProposal confirms and creates the model-proposed deck. Returns ""
// when the user declines (the menu loops).
func createFromProposal(a *app.App, proposal *ai.DeckProposal, from, to string) (string, error) {
	name, lfrom, lto, err := resolveNewDeck(proposal, from, to)
	if err != nil {
		return "", err
	}
	return createDeck(a, name, lfrom, lto)
}

// createFromName prompts for a new deck name, then creates it. Returns ""
// when the input was empty (the menu loops).
func createFromName(a *app.App, proposal *ai.DeckProposal, from, to string) (string, error) {
	line, err := promptReadLine("New deck name: ", nil)
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(line)
	if name == "" {
		fmt.Println("Deck name cannot be empty.")
		return "", nil
	}
	lfrom, lto, err := resolveLanguages(proposal, from, to)
	if err != nil {
		return "", err
	}
	return createDeck(a, name, lfrom, lto)
}

// resolveNewDeck confirms the proposed name and resolves the language pair for
// a deck creation. If there is no proposal, it falls back to manual naming.
func resolveNewDeck(proposal *ai.DeckProposal, from, to string) (string, string, string, error) {
	name := ""
	var err error
	if proposal != nil {
		name = proposal.Name
	} else {
		if line, e := promptReadLine("New deck name: ", nil); e != nil {
			return "", "", "", e
		} else {
			name = strings.TrimSpace(line)
		}
		if name == "" {
			fmt.Println("Deck name cannot be empty.")
			return "", "", "", nil
		}
	}

	lfrom, lto, err := resolveLanguages(proposal, from, to)
	if err != nil {
		return "", "", "", err
	}

	ok, err := promptYesNo(fmt.Sprintf("Create new deck %q (%s \u2192 %s)?", name, lfrom, lto))
	if err != nil {
		return "", "", "", err
	}
	if !ok {
		fmt.Println("Creation declined; choose another option.")
		return "", "", "", nil
	}
	return name, lfrom, lto, nil
}

// resolveLanguages returns the language pair for a new deck: -F/-T flags win,
// then the model proposal, then a prompt.
func resolveLanguages(proposal *ai.DeckProposal, from, to string) (string, string, error) {
	if from == "" && proposal != nil {
		from = proposal.Language
	}
	if to == "" && proposal != nil {
		to = proposal.TranslationLanguage
	}
	if from == "" {
		if line, err := promptReadLine("Source language (e.g. fr): ", nil); err != nil {
			return "", "", err
		} else {
			from = strings.TrimSpace(line)
		}
	}
	if to == "" {
		if line, err := promptReadLine("Translation language (e.g. en): ", nil); err != nil {
			return "", "", err
		} else {
			to = strings.TrimSpace(line)
		}
	}
	if from == "" || to == "" {
		return "", "", fmt.Errorf("a language pair is needed to create a deck")
	}
	return from, to, nil
}

// createDeck writes a new empty deck via the existing create path.
func createDeck(a *app.App, name, from, to string) (string, error) {
	if err := (&CreateCmd{Deck: name, From: from, To: to}).Run(a); err != nil {
		return "", err
	}
	return name, nil
}

// selectDeckID lists the existing decks and reads a selection (number, deck id,
// or name; tab completes deck ids). Returns "" when the user backs out.
func selectDeckID(a *app.App, decks []storage.DeckSummary) (string, error) {
	fmt.Println("Existing decks:")
	for i, d := range decks {
		fmt.Printf("  %d. %s (%s) %s \u2192 %s, %d entries\n",
			i+1, d.Name, d.ID, d.Language, d.TranslationLanguage, d.EntryCount)
	}
	completions := make([]string, len(decks))
	for i, d := range decks {
		completions[i] = d.ID
	}

	for {
		input, err := promptReadLine("Select a deck (tab to autocomplete): ", completions)
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return "", nil
		}
		if n, err := strconv.Atoi(input); err == nil && n >= 1 && n <= len(decks) {
			return decks[n-1].ID, nil
		}
		for _, d := range decks {
			if d.ID == input || d.Name == input {
				return d.ID, nil
			}
		}
		fmt.Printf("Unknown deck %q. Pick a number or a deck id.\n", input)
	}
}

// errIsAbort reports whether err is the interactive-cancel sentinel.
func errIsAbort(err error) bool {
	return errors.Is(err, errAborted)
}
