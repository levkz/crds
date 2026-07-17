package ui

import "github.com/charmbracelet/lipgloss"

var Theme = struct {
	Primary lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Danger  lipgloss.Style
	Muted   lipgloss.Style
	Header  lipgloss.Style
}{
	Primary: lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")),

	Success: lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")),

	Warning: lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")),

	Danger: lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")),

	Muted: lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")),

	Header: lipgloss.NewStyle().
		Bold(true),
}
