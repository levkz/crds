package cli

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"crds/internal/ai"
	"crds/internal/model"
	"crds/internal/ui"
)

type fakeAIClient struct {
	reply       string
	replies     []string
	system, user string
	systems, users []string
}

func (f *fakeAIClient) Complete(_ context.Context, system, user string) (string, error) {
	f.system = system
	f.user = user
	f.systems = append(f.systems, system)
	f.users = append(f.users, user)
	if len(f.replies) > 0 {
		r := f.replies[0]
		f.replies = f.replies[1:]
		return r, nil
	}
	return f.reply, nil
}

func stubAIClient(t *testing.T, reply string) *fakeAIClient {
	t.Helper()
	return stubAIClientInstance(t, &fakeAIClient{reply: reply})
}

func stubAIClientInstance(t *testing.T, fake *fakeAIClient) *fakeAIClient {
	t.Helper()
	orig := resolveAIClient
	resolveAIClient = func() (ai.Client, error) {
		return fake, nil
	}
	t.Cleanup(func() { resolveAIClient = orig })
	return fake
}

func feedStdin(t *testing.T, text string) {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(text); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	orig := os.Stdin
	os.Stdin = file
	t.Cleanup(func() {
		os.Stdin = orig
		_ = file.Close()
	})
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

const fillReply = `- id: hoy
  term: hoy
  translations:
    - text: today
  notes: from ai
  tags:
    - common
`

func TestAiInterpretCmd_Run(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)

	fake := stubAIClient(t, fillReply)
	cmd := &AiInterpretCmd{Deck: "spanish", Text: "hola = hello"}

	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiInterpretCmd.Run: %v", err)
		}
	})
	if fake.system == "" || fake.user == "" {
		t.Fatal("fake client was not called with prompts")
	}
	if !strings.Contains(out, "hoy") || !strings.Contains(out, "today") {
		t.Fatalf("expected interpreted entries in output, got:\n%s", out)
	}
}

func TestAiInterpretCmd_Run_DryRun(t *testing.T) {
	a := newTestApp(t)
	fake := stubAIClient(t, "")

	cmd := &AiInterpretCmd{Text: "hola = hello", DryRun: true}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiInterpretCmd.Run (dry): %v", err)
	}
	if fake.system != "" {
		t.Fatal("dry run must not call the client")
	}
}

func TestAiFillCmd_Run_DryRun(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)
	fake := stubAIClient(t, "")

	cmd := &AiFillCmd{Deck: "spanish", Text: "term: hola\ntranslations:\n  - text: salutation", DryRun: true}
	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiFillCmd.Run (dry): %v", err)
		}
	})
	if fake.system != "" {
		t.Fatal("dry run must not call the client")
	}
	if !strings.Contains(out, "target language: fr") {
		t.Fatalf("expected deck context in prompt, got:\n%s", out)
	}
}

const emptyFieldsReply = `- id: ""
  term: hola
  translations:
    - text: hello
  examples: []
  tags: []
  notes: ""
`

func TestAiInterpretCmd_MinimalRender(t *testing.T) {
	a := newTestApp(t)
	stubAIClient(t, emptyFieldsReply)

	cmd := &AiInterpretCmd{Text: "hola"}
	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiInterpretCmd.Run: %v", err)
		}
	})
	for _, unexp := range []string{"id:", "examples:", "tags:", "notes:"} {
		if strings.Contains(out, unexp) {
			t.Fatalf("minimal output should not contain %q, got:\n%s", unexp, out)
		}
	}
	if !strings.Contains(out, "term: hola") || !strings.Contains(out, "text: hello") {
		t.Fatalf("expected term + translations in output, got:\n%s", out)
	}
}

func TestAiInterpretCmd_Full(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)
	fake := stubAIClient(t, "- term: hola\n  translations:\n    - text: hello\n  examples:\n    - text: Hola, como estas?\n      translation: Hello, how are you?\n  tags:\n    - common\n")

	cmd := &AiInterpretCmd{Deck: "spanish", Text: "hola", Full: true}
	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiInterpretCmd.Run (full): %v", err)
		}
	})
	if !strings.Contains(fake.system, "at least 4") {
		t.Fatalf("full mode should use the full prompt, got:\n%s", fake.system)
	}
	if !strings.Contains(fake.user, "allowed theme tags: common") {
		t.Fatalf("full mode with a deck should pass the deck tags as theme allowlist, got:\n%s", fake.user)
	}
	if !strings.Contains(fake.user, "structural tags are always allowed") {
		t.Fatalf("full mode should remind the model structural tags are allowed, got:\n%s", fake.user)
	}
	if !strings.Contains(out, "Hola, como estas?") {
		t.Fatalf("expected examples in full output, got:\n%s", out)
	}
}

func TestAiInterpretCmd_FullNoDeckAllowsStructuralTags(t *testing.T) {
	a := newTestApp(t)
	fake := stubAIClient(t, "- term: hola\n  translations:\n    - text: hello\n")

	cmd := &AiInterpretCmd{Text: "hola", Full: true, TranslateFrom: "es", TranslateTo: "en"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiInterpretCmd.Run: %v", err)
	}
	if !strings.Contains(fake.user, "allowed theme tags: none (add structural tags and a concise theme tag)") {
		t.Fatalf("full mode without a deck should allow structural tags, got:\n%s", fake.user)
	}
}

func TestAiInterpretCmd_Msg(t *testing.T) {
	a := newTestApp(t)
	fake := stubAIClient(t, fillReply)

	cmd := &AiInterpretCmd{Text: "hola", Msg: "use formal register"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiInterpretCmd.Run: %v", err)
	}
	if !strings.Contains(fake.user, "use formal register") {
		t.Fatalf("extra instruction should be sent to the model, got:\n%s", fake.user)
	}
}

func TestAiInterpretCmd_MinimalAndFullConflict(t *testing.T) {
	a := newTestApp(t)
	cmd := &AiInterpretCmd{Text: "hola", Minimal: true, Full: true}
	if err := cmd.Run(a); err == nil {
		t.Fatal("expected error when both --minimal and --full are set")
	}
}

func TestAiInterpretCmd_TranslateFlags(t *testing.T) {
	a := newTestApp(t)
	fake := stubAIClient(t, fillReply)

	cmd := &AiInterpretCmd{Text: "hola", TranslateFrom: "es", TranslateTo: "de"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiInterpretCmd.Run: %v", err)
	}
	if !strings.Contains(fake.user, "Source language: es") {
		t.Fatalf("expected source language es, got:\n%s", fake.user)
	}
	if !strings.Contains(fake.user, "Target language for translations: de") {
		t.Fatalf("expected target language de, got:\n%s", fake.user)
	}
}

func TestAiInterpretCmd_FullTranslateDryRun(t *testing.T) {
	a := newTestApp(t)
	stubAIClient(t, "")

	cmd := &AiInterpretCmd{Text: "hola", Full: true, TranslateFrom: "es", TranslateTo: "de", DryRun: true}
	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiInterpretCmd.Run (full dry): %v", err)
		}
	})
	if !strings.Contains(out, "source language: es") {
		t.Fatalf("expected source language es, got:\n%s", out)
	}
	if !strings.Contains(out, "target language: de") {
		t.Fatalf("expected target language de, got:\n%s", out)
	}
}

func TestAiFillCmd_TranslateOverride(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)
	stubAIClient(t, "")

	// Deck is en->fr; flags must override to fr->en.
	cmd := &AiFillCmd{Deck: "spanish", Text: "term: hola\ntranslations:\n  - text: salutation", TranslateFrom: "fr", TranslateTo: "en", DryRun: true}
	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiFillCmd.Run (dry): %v", err)
		}
	})
	if !strings.Contains(out, "source language: fr") {
		t.Fatalf("expected source language fr (override), got:\n%s", out)
	}
	if !strings.Contains(out, "target language: en") {
		t.Fatalf("expected target language en (override), got:\n%s", out)
	}
}

func TestAiFillCmd_Msg(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)
	fake := stubAIClient(t, fillReply)

	cmd := &AiFillCmd{Deck: "spanish", Text: "term: hola\ntranslations:\n  - text: salutation", Msg: "add IPA"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiFillCmd.Run: %v", err)
	}
	if !strings.Contains(fake.user, "add IPA") {
		t.Fatalf("extra instruction should be sent to the model, got:\n%s", fake.user)
	}
}

func TestAiAddCmd_Msg(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "swedish")
	syncDecks(t, a)
	fake := stubAIClient(t, fillReply)
	feedStdin(t, "d\n")

	cmd := &AiAddCmd{Deck: "swedish", Text: "hej wait", Msg: "avoid casual slang"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}
	if !strings.Contains(fake.user, "avoid casual slang") {
		t.Fatalf("extra instruction should be sent to the model (last call), got:\n%s", fake.user)
	}
}

func TestAiAddCmd_Run_Append(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "polish")
	syncDecks(t, a)

	stubAIClient(t, fillReply)
	feedStdin(t, "a\n")

	cmd := &AiAddCmd{Deck: "polish", Text: "hoy wait"} // non-structured → interpret + fill
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}

	deck, err := a.Store.LoadDeck("polish")
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}
	found := false
	for _, card := range deck.Cards {
		if card.ID == "hoy" {
			found = true
		}
	}
	if !found {
		t.Fatal("entry 'hoy' was not appended to the deck")
	}
}

func TestAiAddCmd_Run_Discard(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)

	stubAIClient(t, fillReply)
	feedStdin(t, "d\n")

	cmd := &AiAddCmd{Deck: "spanish", Text: "whatever"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}

	deck, err := a.Store.LoadDeck("spanish")
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}
	for _, card := range deck.Cards {
		if card.ID == "hoy" {
			t.Fatal("entry 'hoy' should not be in deck after discard")
		}
	}
}

func TestAiAddCmd_NoDeck_ConfirmSuggested(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)

	fake := stubAIClientInstance(t, &fakeAIClient{
		replies: []string{`{"deck": "spanish"}`, fillReply, fillReply},
	})
	feedStdin(t, "y\na\n")

	cmd := &AiAddCmd{Text: "hoy wait"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}
	if !strings.Contains(strings.Join(fake.users, "\n"), "id: spanish") {
		t.Fatalf("deck suggestion should receive the deck list, got users:\n%s", strings.Join(fake.users, "\n"))
	}

	deck, err := a.Store.LoadDeck("spanish")
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}
	if !hasCard(deck, "hoy") {
		t.Fatal("entry 'hoy' was not appended to the confirmed deck")
	}
}

func TestAiAddCmd_NoDeck_RejectThenSelectExisting(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	writeTestDeck(t, a.DataDir, "french")
	syncDecks(t, a)

	// Model suggests spanish; user declines, selects french by name.
	stubAIClientInstance(t, &fakeAIClient{
		replies: []string{`{"deck": "spanish"}`, fillReply, fillReply},
	})
	feedStdin(t, "n\ns\nfrench\na\n")

	cmd := &AiAddCmd{Text: "hoy wait"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}

	spanish, err := a.Store.LoadDeck("spanish")
	if err != nil {
		t.Fatalf("LoadDeck(spanish): %v", err)
	}
	if hasCard(spanish, "hoy") {
		t.Fatal("entry 'hoy' should not be appended to the declined deck spanish")
	}
	french, err := a.Store.LoadDeck("french")
	if err != nil {
		t.Fatalf("LoadDeck(french): %v", err)
	}
	if !hasCard(french, "hoy") {
		t.Fatal("entry 'hoy' was not appended to the selected deck french")
	}
}

func TestAiAddCmd_NoDeck_CreateProposed(t *testing.T) {
	a := newTestApp(t)

	stubAIClientInstance(t, &fakeAIClient{
		replies: []string{
			`{"deck": null, "proposed": {"name": "French Basics", "from": "fr", "to": "en"}}`,
			fillReply,
			fillReply,
		},
	})
	feedStdin(t, "c\ny\na\n")

	cmd := &AiAddCmd{Text: "hoy wait"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}

	deck, err := loadDeckModel(a, "French Basics")
	if err != nil {
		t.Fatalf("loadDeckModel: %v", err)
	}
	if deck.Language != "fr" || deck.TranslationLanguage != "en" {
		t.Fatalf("deck languages = %s -> %s, want fr -> en", deck.Language, deck.TranslationLanguage)
	}
	if !hasCardByID(deck, "hoy") {
		t.Fatal("entry 'hoy' was not appended to the created deck")
	}
}

func TestAiAddCmd_NoDeck_Abort(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)

	stubAIClientInstance(t, &fakeAIClient{
		replies: []string{`{"deck": "spanish"}`},
	})
	feedStdin(t, "n\na\n")

	cmd := &AiAddCmd{Text: "hoy wait"}
	if err := cmd.Run(a); err != nil {
		t.Fatalf("AiAddCmd.Run: %v", err)
	}

	deck, err := a.Store.LoadDeck("spanish")
	if err != nil {
		t.Fatalf("LoadDeck: %v", err)
	}
	if hasCard(deck, "hoy") {
		t.Fatal("entry 'hoy' should not be appended after abort")
	}
}

func TestAiFillCmd_NoDeck_DryRunErrors(t *testing.T) {
	a := newTestApp(t)
	stubAIClient(t, "")

	cmd := &AiFillCmd{Text: "term: hola\ntranslations:\n  - text: hello", DryRun: true}
	if err := cmd.Run(a); err == nil {
		t.Fatal("expected error for --dry-run without a deck")
	}
}

func TestAiFillCmd_NoDeck_ResolvesAndFills(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "spanish")
	syncDecks(t, a)

	stubAIClientInstance(t, &fakeAIClient{
		replies: []string{`{"deck": "spanish"}`, fillReply},
	})
	feedStdin(t, "y\n")

	cmd := &AiFillCmd{Text: "term: hola\ntranslations:\n  - text: hello"}
	out := captureStdout(t, func() {
		if err := cmd.Run(a); err != nil {
			t.Fatalf("AiFillCmd.Run: %v", err)
		}
	})
	if !strings.Contains(out, "hoy") || !strings.Contains(out, "today") {
		t.Fatalf("expected filled entries in output, got:\n%s", out)
	}
}

func TestTagListCmd_DeckWide(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "greek")
	syncDecks(t, a)

	out := captureStdout(t, func() {
		if err := (&TagListCmd{Deck: "greek"}).Run(a); err != nil {
			t.Fatalf("TagListCmd.Run (deck-wide): %v", err)
		}
	})
	if !strings.Contains(out, "common") {
		t.Fatalf("expected deck-wide tag 'common', got:\n%s", out)
	}
}

func TestTagListCmd_SingleTerm(t *testing.T) {
	a := newTestApp(t)
	writeTestDeck(t, a.DataDir, "greek")
	syncDecks(t, a)

	// e2 has tag "common"; e1 has none.
	out := captureStdout(t, func() {
		if err := (&TagListCmd{Deck: "greek", TermID: "e1"}).Run(a); err != nil {
			t.Fatalf("TagListCmd.Run (term): %v", err)
		}
	})
	if !strings.Contains(out, "No tags") {
		t.Fatalf("expected 'No tags' message, got:\n%s", out)
	}
}

func hasCard(deck ui.DeckData, id string) bool {
	for _, card := range deck.Cards {
		if card.ID == id {
			return true
		}
	}
	return false
}

func hasCardByID(deck *model.Deck, id string) bool {
	for _, e := range deck.Entries {
		if e.ID == id {
			return true
		}
	}
	return false
}

