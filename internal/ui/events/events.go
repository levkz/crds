package events

import "time"

// TickMsg is sent on a 1-second heartbeat.
type TickMsg time.Time

// ThemeSwitchMsg requests a theme change.
type ThemeSwitchMsg struct {
	Name string
}

// ShowNotificationMsg displays a transient notification.
type ShowNotificationMsg struct {
	Text string
}

// HideNotificationMsg clears the current notification.
type HideNotificationMsg struct{}
