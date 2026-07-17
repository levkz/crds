package model

import "time"

type Progress struct {
	DeckID  string
	EntryID string

	Ease     float64
	Interval int

	Due time.Time

	Correct   int
	Incorrect int
}
