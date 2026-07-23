// Package fuzzy provides string matching utilities with pluggable similarity algorithms.
package fuzzy

import (
	"strings"
)

func Normalize(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(text)), " ")
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
	Again = 1
	Hard  = 2
	Good  = 3
)

type FuzzyMatcher struct {
	Matcher
	Threshold float64
}

func NewFuzzyMatcher(threshold float64) *FuzzyMatcher {
	if threshold <= 0 {
		threshold = 0.7
	}
	return &FuzzyMatcher{
		Matcher:   &LevenshteinMatcher{},
		Threshold: threshold,
	}
}

func (fm *FuzzyMatcher) Check(input string, correctAnswers []string) (bestScore float64, matched bool) {
	input = Normalize(input)
	for _, answer := range correctAnswers {
		norm := Normalize(answer)
		score := fm.Similarity(input, norm)
		if score > bestScore {
			bestScore = score
		}
	}
	matched = bestScore >= fm.Threshold
	return
}

func (fm *FuzzyMatcher) Grade(input string, correctAnswers []string) int {
	input = Normalize(input)
	bestScore := 0.0
	for _, answer := range correctAnswers {
		norm := Normalize(answer)
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
