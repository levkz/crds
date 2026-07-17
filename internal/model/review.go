package model

import "time"

type Review struct {
	DeckID     string
	EntryID    string
	ReviewedAt time.Time
	Grade      int
}
