package cli

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"crds/internal/ai"
)

type fakeAIClient struct {
	reply       string
	system, user string
}

func (f *fakeAIClient) Complete(_ context.Context, system, user string) (string, error) {
	f.system = system
	f.user = user
	return f.reply, nil
}

func stubAIClient(t *testing.T, reply string) *fakeAIClient {
	t.Helper()
	fake := &fakeAIClient{reply: reply}
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
	if !strings.Contains(out, "Hola, como estas?") {
		t.Fatalf("expected examples in full output, got:\n%s", out)
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

