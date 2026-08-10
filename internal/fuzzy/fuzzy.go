// Package fuzzy provides string matching utilities with pluggable similarity algorithms.
package fuzzy

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Mode selects how typed answers are compared to correct answers.
type Mode int

const (
	// Strict compares answers exactly (accents matter).
	Strict Mode = iota
	// Approximate strips accents from both sides before comparing, so
	// "cafe" matches "café" exactly.
	Approximate
)

// ParseMode parses a mode name. An empty string yields the default
// (Approximate). Unknown names also fall back to the default.
func ParseMode(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return Strict
	default:
		return Approximate
	}
}

func (m Mode) String() string {
	switch m {
	case Strict:
		return "strict"
	default:
		return "approximate"
	}
}

func Normalize(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
}

// StripAccents removes combining diacritical marks (Unicode category Mn) by
// decomposing to NFD, dropping marks, and recomposing to NFC. Ligatures and
// letters that do not NFD-decompose (e.g. ß, ø, ł, å) are left untouched.
func StripAccents(text string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(text) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return norm.NFC.String(b.String())
}

type Matcher interface {
	Similarity(a, b string) float64
}

type LevenshteinMatcher struct{}

func (m *LevenshteinMatcher) Similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)

	if la > lb {
		ra, rb = rb, ra
		la, lb = lb, la
	}

	row := make([]int, la+1)
	for i := 1; i <= la; i++ {
		row[i] = i
	}

	for j := 1; j <= lb; j++ {
		prev := row[0]
		row[0] = j
		for i := 1; i <= la; i++ {
			cur := row[i]
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			row[i] = min(prev+cost, row[i-1]+1, row[i]+1)
			prev = cur
		}
	}

	dist := float64(row[la])
	maxLen := float64(lb)
	return 1.0 - dist/maxLen
}

const (
	Again = 0
	Hard  = 1
	Good  = 2
)

type FuzzyMatcher struct {
	Matcher
	Threshold float64
	// Mode selects accent handling. Approximate strips accents from both
	// sides before comparing; Strict leaves them intact.
	Mode Mode
}

func NewFuzzyMatcher(threshold float64) *FuzzyMatcher {
	if threshold <= 0 {
		threshold = 0.7
	}
	return &FuzzyMatcher{
		Matcher:   &LevenshteinMatcher{},
		Threshold: threshold,
		Mode:      Approximate,
	}
}

func (fm *FuzzyMatcher) prepare(text string) string {
	text = Normalize(text)
	if fm.Mode == Approximate {
		text = StripAccents(text)
	}
	return text
}

func (fm *FuzzyMatcher) Check(input string, correctAnswers []string) (bestScore float64, matched bool) {
	input = fm.prepare(input)
	for _, answer := range correctAnswers {
		norm := fm.prepare(answer)
		score := fm.Similarity(input, norm)
		if score > bestScore {
			bestScore = score
		}
	}
	matched = bestScore >= fm.Threshold
	return
}

func (fm *FuzzyMatcher) Grade(input string, correctAnswers []string) int {
	input = fm.prepare(input)
	bestScore := 0.0
	for _, answer := range correctAnswers {
		norm := fm.prepare(answer)
		if input == norm {
			return Good
		}
		score := fm.Similarity(input, norm)
		if score > bestScore {
			bestScore = score
		}
	}
	if bestScore >= fm.Threshold {
		return Hard
	}
	return Again
}
