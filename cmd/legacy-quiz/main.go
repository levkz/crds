package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Entry struct {
	Term            string
	TranslationText string
	Translations    []string
	Description     string
}

var normalizer = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeAnswer(text string) string {
	text = strings.ToLower(text)
	text = normalizer.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func parseLine(line string) Entry {
	parts := strings.Split(line, "=>")

	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	entry := Entry{
		Term: parts[0],
	}

	if len(parts) > 1 {
		entry.TranslationText = parts[1]

		for _, t := range strings.Split(parts[1], "/") {
			t = strings.TrimSpace(t)
			if t != "" {
				entry.Translations = append(entry.Translations, t)
			}
		}

		if len(entry.Translations) == 0 && entry.TranslationText != "" {
			entry.Translations = []string{entry.TranslationText}
		}
	}

	if len(parts) > 2 {
		entry.Description = strings.TrimSpace(strings.Join(parts[2:], "=>"))
	}

	return entry
}

func shuffle(entries []Entry) []Entry {
	shuffled := slices.Clone(entries)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func getFilename() string {
	if len(os.Args) > 1 {
		return filepath.Join("exercises", os.Args[1]+".txt")
	}

	files, err := os.ReadDir("exercises")
	if err != nil {
		panic(err)
	}

	var names []string
	for _, file := range files {
		if before, ok := strings.CutSuffix(file.Name(), ".txt"); ok {
			names = append(names, before)
		}
	}

	fmt.Println("Available files:")
	for i, name := range names {
		fmt.Printf("- %d. %s\n", i+1, name)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the name or number of the file: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if n, err := strconv.Atoi(input); err == nil {
		if n >= 1 && n <= len(names) {
			input = names[n-1]
		}
	}

	found := slices.Contains(names, input)
	if !found {
		fmt.Printf("File %q not found. Using first file.\n", input)
		input = names[0]
	}

	return filepath.Join("exercises", input+".txt")
}

func ignoreComments(lines []string) []string {
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		result = append(result, line)
	}

	return result
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func prompt(reader *bufio.Reader, message string) string {
	fmt.Print(message)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func main() {
	filename := getFilename()

	lines, err := readLines(filename)
	if err != nil {
		panic(err)
	}

	lines = ignoreComments(lines)

	var original []Entry
	for _, line := range lines {
		original = append(original, parseLine(line))
	}

	active := slices.Clone(original)
	skipped := map[string]bool{}

	reader := bufio.NewReader(os.Stdin)

	round := 1

	const (
		purple = "\033[95m"
		reset  = "\033[0m"
	)

	for len(active) > 0 {
		fmt.Printf("\n--- Round %d ---\n", round)

		active = shuffle(active)

		var remaining []Entry

		for i, entry := range active {

			fmt.Printf("\n[%d/%d] 🇫🇷 %s\n", i+1, len(active), entry.Term)

			input := strings.ToLower(
				prompt(reader, "(Guess the translation or press ↵ to reveal) "),
			)

			if input == "s" {
				input = ""
			}

			valid := map[string]struct{}{}

			for _, t := range entry.Translations {
				valid[normalizeAnswer(t)] = struct{}{}
			}

			if entry.TranslationText != "" {
				valid[normalizeAnswer(entry.TranslationText)] = struct{}{}
			}

			known := false

			if input != "" {
				if _, ok := valid[normalizeAnswer(input)]; ok {
					fmt.Println("✅ Correct guess!")
					skipped[entry.Term] = true
					known = true
				} else {
					fmt.Println("❌ Not quite.")
				}
			} else {
				fmt.Println("...")
			}

			display := entry.TranslationText
			if display == "" {
				display = "[No translation]"
			}

			fmt.Printf("[%d/%d] 🇫🇷 %s => 🇬🇧 %s",
				i+1,
				len(active),
				entry.Term,
				display,
			)

			if entry.Description != "" {
				fmt.Printf(" => %s%s%s", purple, entry.Description, reset)
			}

			fmt.Println()

			if !known {
				answer := strings.ToLower(
					prompt(reader, "Did you know it? Press 's' to skip next time, ↵ to keep reviewing: "),
				)

				if answer == "s" {
					skipped[entry.Term] = true
				}
			}

			remaining = append(remaining, entry)
		}

		active = active[:0]
		for _, entry := range remaining {
			if !skipped[entry.Term] {
				active = append(active, entry)
			}
		}

		round++
	}

	fmt.Println("\n✅ All words reviewed or skipped. Goodbye!")
}
