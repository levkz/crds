package theme

import "github.com/charmbracelet/lipgloss"

type Borders struct {
	Normal  lipgloss.Border
	Rounded lipgloss.Border
	Double  lipgloss.Border
	Thick   lipgloss.Border
	None    lipgloss.Border
}

var DefaultBorders = Borders{
	Normal:  lipgloss.NormalBorder(),
	Rounded: lipgloss.RoundedBorder(),
	Double:  lipgloss.DoubleBorder(),
	Thick:   lipgloss.ThickBorder(),
	None:    lipgloss.HiddenBorder(),
}
