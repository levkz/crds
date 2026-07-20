package layout

import "github.com/charmbracelet/lipgloss"

func Stack(layers ...string) string {
	if len(layers) == 0 {
		return ""
	}
	if len(layers) == 1 {
		return layers[0]
	}

	base := layers[len(layers)-1]
	over := layers[len(layers)-2]
	w := lipgloss.Width(base)
	h := lipgloss.Height(base)
	placed := lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, over)
	result := lipgloss.JoinVertical(lipgloss.Top, placed, base)

	if len(layers) > 2 {
		return Stack(append(layers[:len(layers)-2], result)...)
	}
	return result
}
