package ai

import (
	"fmt"
	"strings"

	"crds/internal/model"

	"go.yaml.in/yaml/v3"
)

// LanguageContext carries the deck languages used when no deck context exists
// (the interpreter can run without a deck).
type LanguageContext struct {
	Language            string
	TranslationLanguage string
}

// DeckContext is everything the filler agent needs beyond the entries
// themselves: languages, the tag allowlist, and style samples.
type DeckContext struct {
	Language            string
	TranslationLanguage string
	AllowedTags         []string
	Samples             []model.Entry
}

// termConventions is shared by the fill agent and the full-effort interpreter.
const termConventions = `Term and translation conventions (use CRDS variant syntax):
- Optional text goes in parentheses (...); required alternatives go in square brackets [...].
- French feminine nouns: prefix the article, e.g. "baguette" -> "(une/la) baguette"; before a vowel sound use the elided form "(une /l') baguette".
- French masculine nouns: "chat" -> "(un/le) chat"; before a vowel sound use "(un /l') chat".
- Verbs: the English translation is prepended with "(to)", e.g. "eat" -> "(to) eat".
- Conjugated verb stems: use required alternatives, e.g. "mang[er/e/ons]" or "mang[er/ez/e/ons/ent]".
- Adjectives with gender agreement: "série[ux/use]", "bavard(e/s/es)".
- Match the notation style of the sample entries when provided.`

// tagRules is shared by the fill agent and the full-effort interpreter.
// Structural tags are always allowed; theme tags come from the deck allowlist
// (or a concise model-chosen theme tag when none is supplied).
const tagRules = `- Structural tags are ALWAYS allowed and must be added when relevant: noun, verb, adjective, adverb, pronoun, conjunction, preposition; gender tags (masculin, feminin, and neutral — neutral only if the language has grammatical gender); and verb class where the language uses one (e.g. French 1st/2nd/3rd group, Spanish -ar/-er/-ir).
- Proficiency: add exactly one CEFR level tag (A1, A2, B1, B2, C1, C2) estimating the term's difficulty.
- Theme tags: choose ONLY from the allowed theme tags in the user message. If none are listed, add a concise theme tag that fits the entry (e.g. greetings, food, travel).`

// InterpretMessages builds the system and user messages for the interpreter
// agent: unstructured free text -> minimal structured YAML entries. msg is an
// optional extra instruction passed through to the model.
func InterpretMessages(raw string, lc LanguageContext, msg string) (string, string) {
	system := `You convert free-form vocabulary lists into structured YAML entries for a flashcard application.

Rules:
- Extract every distinct word or phrase the user listed; merge duplicates.
- Each entry is a YAML mapping with these fields only:
  term: the word or phrase in the source language
  translations: a list of { text: <translation> }
  Keep tags, notes, and examples ONLY when the user supplied them.
- Do not invent example sentences, tags, or notes.
- Preserve the user's spelling and punctuation.
- Output ONLY a YAML sequence of entries. No prose, no explanations, no markdown fences.`

	var b strings.Builder
	b.WriteString("Convert the following into YAML entries.\n\n")
	if lc.Language != "" {
		fmt.Fprintf(&b, "Source language: %s\n", lc.Language)
	}
	if lc.TranslationLanguage != "" {
		fmt.Fprintf(&b, "Target language for translations: %s\n", lc.TranslationLanguage)
	}
	b.WriteString("\nRaw input:\n")
	b.WriteString(raw)
	appendMsg(&b, msg)

	return system, b.String()
}

// InterpretFullMessages builds the system and user messages for full-effort
// interpretation: free text -> complete entries with examples, notes, and
// deck-constrained tags. msg is an optional extra instruction for the model.
func InterpretFullMessages(raw string, dc DeckContext, msg string) (string, string) {
	system := `You convert free-form vocabulary lists into structured, complete YAML entries for a flashcard application that learns words as (<term> -> <translation>).

Rules:
- Extract every distinct word or phrase the user listed; merge duplicates.
- Each entry is a YAML mapping with:
  term: the word or phrase in the source language
  translations: a list of { text: <translation> }
  examples: a list of at least 4 objects, each { text: <a complete, natural sentence in the SOURCE language>, translation: <its translation in the TARGET language> }. Both text and translation must be non-empty.
  notes: only when it adds useful learning context (gender, usage nuance, common error)
  tags: a non-empty list — see the tag rules below
- Do not set the id field.
- Preserve the user's spelling and punctuation.
- Output ONLY a YAML sequence of entries. No prose, no explanations, no markdown fences.

Tag rules:
` + tagRules + `

Example of a correctly formed entry:
- term: hola
  translations:
    - text: hello
  examples:
    - text: Hola, como estas?
      translation: Hello, how are you?
    - text: Dijo hola desde la ventana.
      translation: He said hello from the window.
    - text: Un simple hola puede alegrar el dia.
      translation: A simple hello can brighten the day.

` + termConventions

	var b strings.Builder
	b.WriteString("Deck context:\n")
	if dc.Language != "" {
		fmt.Fprintf(&b, "- source language: %s\n", dc.Language)
	}
	if dc.TranslationLanguage != "" {
		fmt.Fprintf(&b, "- target language: %s\n", dc.TranslationLanguage)
	}
	if len(dc.AllowedTags) > 0 {
		b.WriteString("- allowed theme tags: ")
		b.WriteString(strings.Join(dc.AllowedTags, ", "))
		b.WriteString(" (structural tags are always allowed)\n")
	} else {
		b.WriteString("- allowed theme tags: none (add structural tags and a concise theme tag)\n")
	}

	if len(dc.Samples) > 0 {
		b.WriteString("\nExisting entries in this deck (match their style and conventions):\n")
		for _, sample := range dc.Samples {
			data, err := yaml.Marshal(sample)
			if err == nil {
				fmt.Fprintf(&b, "%s", data)
			}
		}
	}

	b.WriteString("\nRaw input:\n")
	b.WriteString(raw)
	appendMsg(&b, msg)

	return system, b.String()
}

// FillMessages builds the system and user messages for the filler agent:
// (possibly partial) YAML entries -> completed entries. msg is an optional
// extra instruction passed through to the model.
func FillMessages(entries []model.Entry, dc DeckContext, msg string) (string, string) {
	system := `You complete and improve YAML vocabulary entries for a flashcard application that learns words as (<term> -> <translation>).

Rules:
- Keep every field the user supplied; fix only clear typos.
- Never change an existing term's core words or meaning; you may improve its form using the conventions below.
- Add at least 4 examples; each is { text: <a complete, natural sentence in the SOURCE language>, translation: <its translation in the TARGET language> }. Both text and translation must be non-empty. Examples must be natural and correct.
- Add a notes field only when it adds useful learning context (gender, usage nuance, common error). Prefer noun gender and article information when relevant.
- Do not set the id field.
- Output ONLY a YAML sequence of entries matching the input order. No prose, no markdown fences.

Tag rules:
` + tagRules + `

` + termConventions

	var b strings.Builder
	b.WriteString("Deck context:\n")
	fmt.Fprintf(&b, "- source language: %s\n", dc.Language)
	fmt.Fprintf(&b, "- target language: %s\n", dc.TranslationLanguage)

	if len(dc.AllowedTags) > 0 {
		b.WriteString("- allowed theme tags: ")
		b.WriteString(strings.Join(dc.AllowedTags, ", "))
		b.WriteString(" (structural tags are always allowed)\n\n")
	} else {
		b.WriteString("- allowed theme tags: none (add structural tags and a concise theme tag)\n\n")
	}

	if len(dc.Samples) > 0 {
		b.WriteString("Existing entries in this deck (match their style and conventions):\n")
		for _, sample := range dc.Samples {
			data, err := yaml.Marshal(sample)
			if err == nil {
				fmt.Fprintf(&b, "%s", data)
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("Entries to complete:\n")
	for _, entry := range entries {
		data, err := yaml.Marshal(entry)
		if err == nil {
			fmt.Fprintf(&b, "%s", data)
		}
	}
	appendMsg(&b, msg)

	return system, b.String()
}

// appendMsg appends the user's extra instruction to a user prompt when set.
func appendMsg(b *strings.Builder, msg string) {
	if msg != "" {
		fmt.Fprintf(b, "\n\nAdditional instruction:\n%s", msg)
	}
}