package ui

import (
	"math/rand"
	"sort"

	"crds/internal/stats"
)

func SortCards(mode QuizMode, cards []CardData, progress map[string]stats.EntryProgress, due []string) []CardData {
	out := make([]CardData, len(cards))
	copy(out, cards)

	switch mode {
	case QuizModeNormal:
	case QuizModeRandom:
		rand.Shuffle(len(out), func(i, j int) {
			out[i], out[j] = out[j], out[i]
		})
	case QuizModeSmart:
		sort.SliceStable(out, func(i, j int) bool {
			return entryConfidence(out[i].ID, progress) < entryConfidence(out[j].ID, progress)
		})
	case QuizModeKindaSmart:
		sort.SliceStable(out, func(i, j int) bool {
			ci := entryConfidence(out[i].ID, progress) + rand.Float64()*0.1
			cj := entryConfidence(out[j].ID, progress) + rand.Float64()*0.1
			return ci < cj
		})
	case QuizModeDue:
		out = filterDue(out, due)
	}
	return out
}

// filterDue keeps only cards whose entry appears in the due set, ordered by
// the due set's ordering (unseen first, then by due date).
func filterDue(cards []CardData, due []string) []CardData {
	if len(due) == 0 {
		return nil
	}
	byID := make(map[string]CardData, len(cards))
	for _, c := range cards {
		byID[c.ID] = c
	}
	filtered := make([]CardData, 0, len(due))
	seen := make(map[string]bool, len(due))
	for _, id := range due {
		c, ok := byID[id]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		filtered = append(filtered, c)
	}
	return filtered
}

func entryConfidence(id string, progress map[string]stats.EntryProgress) float64 {
	if p, ok := progress[id]; ok {
		return p.Confidence()
	}
	return stats.Confidence(0, 0)
}
