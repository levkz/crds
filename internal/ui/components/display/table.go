package components

import (
	"strings"

	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

func Table(headers []string, rows [][]string, width int) string {
	if len(headers) == 0 || width < 1 {
		return ""
	}

	ncols := len(headers)

	colWidths := make([]int, ncols)
	for i, h := range headers {
		colWidths[i] = renderer.VisibleWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < ncols {
				w := renderer.VisibleWidth(cell)
				if w > colWidths[i] {
					colWidths[i] = w
				}
			}
		}
	}

	totalWidth := ncols - 1
	for _, w := range colWidths {
		totalWidth += w
	}
	if totalWidth > width {
		ratio := float64(width-(ncols-1)) / float64(totalWidth-(ncols-1))
		if ratio < 0.3 {
			ratio = 0.3
		}
		for i := range colWidths {
			colWidths[i] = int(float64(colWidths[i]) * ratio)
			if colWidths[i] < 3 {
				colWidths[i] = 3
			}
		}
	}

	var b strings.Builder

	for i, h := range headers {
		if i > 0 {
			b.WriteString(" ")
		}
		truncated := renderer.Truncate(h, colWidths[i])
		padded := truncated + strings.Repeat(" ", colWidths[i]-renderer.VisibleWidth(truncated))
		b.WriteString(styles.MutedText().Render(padded))
	}
	b.WriteString("\n")

	for i := range headers {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strings.Repeat("─", colWidths[i]))
	}

	for _, row := range rows {
		b.WriteString("\n")
		for i, cell := range row {
			if i >= ncols {
				break
			}
			if i > 0 {
				b.WriteString(" ")
			}
			truncated := renderer.Truncate(cell, colWidths[i])
			padded := truncated + strings.Repeat(" ", colWidths[i]-renderer.VisibleWidth(truncated))
			b.WriteString(padded)
		}
	}

	return b.String()
}
