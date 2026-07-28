package components

import (
	"strings"

	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

func RenderList(items []string, selected int, width int) string {
	return RenderListClipped(items, selected, 0, 0, width)
}

func RenderListClipped(items []string, selected int, offset int, maxItems int, width int) string {
	maxItemWidth := width - 3
	if maxItemWidth < 1 {
		maxItemWidth = 1
	}

	if len(items) == 0 {
		return ""
	}

	visible := items
	relSelected := selected
	showAbove := false
	showBelow := false

	if maxItems > 0 && len(items) > maxItems {
		showAbove = offset > 0

		itemLimit := maxItems
		if showAbove {
			itemLimit--
			if itemLimit < 1 {
				itemLimit = 1
				showAbove = false
			}
		}

		end := offset + itemLimit
		if end > len(items) {
			end = len(items)
		}

		showBelow = end < len(items)
		if showBelow {
			itemLimit--
			if itemLimit < 1 {
				itemLimit = 1
				showBelow = false
			}
			end = offset + itemLimit
			if end > len(items) {
				end = len(items)
			}
			showBelow = end < len(items)
		}

		visible = items[offset:end]
		relSelected = selected - offset
	}

	var b strings.Builder
	first := true

	if showAbove {
		if !first {
			b.WriteString("\n")
		}
		b.WriteString(styles.MutedText().Render("  ↑ more above"))
		first = false
	}

	for i, item := range visible {
		if !first {
			b.WriteString("\n")
		}
		first = false
		truncated := renderer.Truncate(item, maxItemWidth)
		if i == relSelected {
			b.WriteString(styles.SelectedItem().Render(ui.Theme.Icons.Navigate + " " + truncated))
		} else {
			b.WriteString(styles.MutedText().Render("  " + truncated))
		}
	}

	if showBelow {
		if !first {
			b.WriteString("\n")
		}
		b.WriteString(styles.MutedText().Render("  ↓ more below"))
	}

	return b.String()
}
