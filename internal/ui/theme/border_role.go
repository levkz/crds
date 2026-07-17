package theme

import "github.com/charmbracelet/lipgloss"

type BorderRole int

const (
	BorderRoleContainer BorderRole = iota
	BorderRoleCard
	BorderRoleModal
	BorderRoleEmphasis
	BorderRoleSection
	BorderRoleNone
)

var defaultBorderForRole = map[BorderRole]lipgloss.Border{
	BorderRoleContainer: lipgloss.NormalBorder(),
	BorderRoleCard:      lipgloss.RoundedBorder(),
	BorderRoleModal:     lipgloss.RoundedBorder(),
	BorderRoleEmphasis:  lipgloss.DoubleBorder(),
	BorderRoleSection:   lipgloss.ThickBorder(),
	BorderRoleNone:      lipgloss.HiddenBorder(),
}
