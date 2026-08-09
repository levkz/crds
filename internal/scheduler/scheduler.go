// Package scheduler implements spaced repetition for review scheduling.
//
// It owns the learning algorithm only — it never touches the terminal UI,
// storage, or configuration. Progress records (model.Progress) flow in,
// scheduled progress flows out. A different algorithm (FSRS, Leitner) can
// replace SM-2 here without affecting callers.
package scheduler

import (
	"math"
	"time"

	"crds/internal/model"
)

// Grade constants mirror the unified 0-3 review scale (ui.Grade) but are
// plain ints so this package stays free of UI dependencies.
const (
	GradeAgain int = iota // 0 — repeat today
	GradeHard             // 1
	GradeGood             // 2 — correct threshold
	GradeEasy             // 3
)

const (
	// initialEase is the ease factor assigned to a brand-new card.
	initialEase = 2.5

	// minEase is the floor for the ease factor; a card can never drop below it.
	minEase = 1.3

	// lapsePenalty is subtracted from ease when a card is failed (Again).
	lapsePenalty = 0.2

	// easeStep is added (Easy) or subtracted (Hard) on subsequent reviews.
	easeStep = 0.15

	// firstAgain is the interval in days after the first failed review.
	firstAgain = 0

	// firstHard and firstGood are the intervals after the first pass.
	firstHard = 1
	firstGood = 1

	// firstEasy is the interval after the first brilliant review.
	firstEasy = 4
)

// Update computes the next scheduling state for a card given its previous
// state and the review grade. A zero Progress (zero Due) means the card has
// never been reviewed. The returned value carries the updated Ease, Interval,
// Due, and correct/incorrect counters.
func Update(prev model.Progress, grade int, now time.Time) model.Progress {
	next := prev
	if next.Ease == 0 {
		next.Ease = initialEase
	}

	// A fresh card is one that has never been scheduled (no Due yet).
	first := prev.Due.IsZero()

	switch grade {
	case GradeAgain:
		next.Ease = floorEase(next.Ease - lapsePenalty)
		next.Interval = firstAgain
		next.Due = now
		next.Incorrect++

	case GradeHard:
		if !first {
			next.Ease = floorEase(next.Ease - easeStep)
		}
		if first {
			next.Interval = firstHard
		} else {
			next.Interval = maxIvl(1, roundIvl(float64(next.Interval)*1.2))
		}
		next.Due = now.AddDate(0, 0, next.Interval)
		next.Correct++

	case GradeGood:
		if first {
			next.Interval = firstGood
		} else {
			next.Interval = maxIvl(1, roundIvl(float64(next.Interval)*next.Ease))
		}
		next.Due = now.AddDate(0, 0, next.Interval)
		next.Correct++

	case GradeEasy:
		if !first {
			next.Ease += easeStep
		}
		if first {
			next.Interval = firstEasy
		} else {
			next.Interval = maxIvl(1, roundIvl(float64(next.Interval)*next.Ease*1.3))
		}
		next.Due = now.AddDate(0, 0, next.Interval)
		next.Correct++
	}

	return next
}

// IsCorrect reports whether a grade counts as a successful review (>= Good).
func IsCorrect(grade int) bool {
	return grade >= GradeGood
}

func floorEase(e float64) float64 {
	return math.Max(minEase, e)
}

func roundIvl(v float64) int {
	return int(math.Round(v))
}

func maxIvl(a, b int) int {
	if a > b {
		return a
	}
	return b
}