package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ergochat/readline"
	"github.com/mattn/go-isatty"
)

// errAborted signals that the user cancelled an interactive flow (a soft stop,
// not a failure).
var errAborted = errors.New("aborted")

// promptReadLine is the test seam for interactive line input (the AI deck
// resolution flow). Tests replace it to feed canned answers without a terminal.
// The real implementation uses readline (with tab-completion over completions)
// when stdin is a TTY and falls back to a plain buffered scan otherwise.
var promptReadLine = func(prompt string, completions []string) (string, error) {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		return readlinePrompt(prompt, completions)
	}
	fmt.Print(prompt)
	return stdinLine()
}

// readlinePrompt runs an ergochat/readline session supporting tab-completion.
func readlinePrompt(prompt string, completions []string) (string, error) {
	cfg := &readline.Config{
		Prompt:       prompt,
		HistoryLimit: -1,
	}
	if len(completions) > 0 {
		items := make([]*readline.PrefixCompleter, len(completions))
		for i, c := range completions {
			items[i] = readline.PcItem(c)
		}
		cfg.AutoComplete = readline.NewPrefixCompleter(items...)
	}
	rl, err := readline.NewFromConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("init readline: %w", err)
	}
	defer func() { _ = rl.Close() }()

	line, err := rl.ReadLine()
	if err != nil {
		if errors.Is(err, readline.ErrInterrupt) {
			return "", errAborted
		}
		return "", err
	}
	return line, nil
}

// stdinLine reads a single line from stdin, buffering across calls so
// consecutive prompts share one reader. The reader is recreated whenever
// os.Stdin changes (e.g. tests swapping stdin between cases).
func stdinLine() (string, error) {
	if stdinReader == nil || stdinReaderFile != os.Stdin {
		stdinReader = bufio.NewReader(os.Stdin)
		stdinReaderFile = os.Stdin
	}
	line, err := stdinReader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// stdinReader is the shared buffered reader behind the non-TTY fallback; it is
// keyed to the os.Stdin it was created over.
var (
	stdinReader     *bufio.Reader
	stdinReaderFile *os.File
)

// promptYesNo asks a y/n question, looping until the user answers. An empty
// answer counts as yes.
func promptYesNo(prompt string) (bool, error) {
	for {
		answer, err := promptReadLine(prompt+" [y/N] ", nil)
		if err != nil {
			if errors.Is(err, errAborted) {
				return false, nil
			}
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "", "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println("Please answer y or n.")
		}
	}
}
