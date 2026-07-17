package model

import "time"

type Session struct {
	Started time.Time

	Reviewed int

	Correct int

	Incorrect int

	Duration time.Duration
}
