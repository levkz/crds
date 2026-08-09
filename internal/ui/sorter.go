package ui

import (
	"math/rand"
	"sort"

	"crds/internal/stats"
)

func SortCards(mode QuizMode, cards []CardData, progress map[string]stats.EntryProgress) []CardData {
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
	}
	return out
}

func entryConfidence(id string, progress map[string]stats.EntryProgress) float64 {
	if p, ok := progress[id]; ok {
		return p.Confidence()
	}
	return stats.Confidence(0, 0)
}
