package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"crds/internal/ui"
	"crds/internal/ui/renderer"
	"crds/internal/ui/styles"
)

func renderQuizTags(tags []string) string {
	var styled []string
	tagStyle := styles.PrimaryBg().Padding(0, 1)
	for _, t := range tags {
		styled = append(styled, tagStyle.Render(t))
	}
	return strings.Join(styled, " ")
}

func renderQuizExamplesBlock(examples []ui.ExampleData, width, height, examplesPage int, topBodyLines int) string {
	pp := quizExamplesPerPage(width, height, topBodyLines)
	if pp <= 0 {
		pp = 1
	}
	start := examplesPage * pp
	if start >= len(examples) {
		return ""
	}
	end := start + pp
	if end > len(examples) {
		end = len(examples)
	}
	page := examples[start:end]

	if width > 80 {
		return renderQuizExamplesTwoCol(page, width)
	}
	return renderQuizExamplesSingleCol(page, width)
}

func renderQuizExamplesSingleCol(examples []ui.ExampleData, width int) string {
	colWidth := width - 2
	if colWidth < 10 {
		colWidth = 10
	}
	var blocks []string
	for _, ex := range examples {
		var blockLines []string
		blockLines = append(blockLines, renderer.Wrap("- "+ex.Text, colWidth)...)
		if ex.Translation != "" {
			blockLines = append(blockLines, renderer.Wrap("  "+ex.Translation, colWidth)...)
		}
		blocks = append(blocks, strings.Join(blockLines, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

func renderQuizExamplesTwoCol(examples []ui.ExampleData, width int) string {
	colWidth := (width - 3) / 2
	if colWidth < 10 {
		colWidth = 10
	}

	var rows []string
	for i := 0; i < len(examples); i += 2 {
		left := renderQuizExampleCell(examples[i], colWidth)
		var right []string
		if i+1 < len(examples) {
			right = renderQuizExampleCell(examples[i+1], colWidth)
		}
		maxH := len(left)
		if len(right) > maxH {
			maxH = len(right)
		}
		for len(left) < maxH {
			left = append(left, strings.Repeat(" ", colWidth))
		}
		for len(right) < maxH {
			right = append(right, strings.Repeat(" ", colWidth))
		}
		for j := 0; j < maxH; j++ {
			rows = append(rows, left[j]+" "+right[j])
		}
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func renderQuizExampleCell(ex ui.ExampleData, width int) []string {
	var lines []string
	lines = append(lines, renderer.Wrap("- "+ex.Text, width)...)
	if ex.Translation != "" {
		lines = append(lines, renderer.Wrap("  "+ex.Translation, width)...)
	}
	for i, l := range lines {
		if w := renderer.VisibleWidth(l); w < width {
			lines[i] = l + strings.Repeat(" ", width-w)
		}
	}
	return lines
}

func quizExamplesPerPage(width, height int, topBodyLines int) int {
	availLines := height - topBodyLines - 3

	if availLines < 3 {
		return 1
	}

	if width > 80 {
		perRow := 2
		linesPerItem := 2
		itemsPerPage := (availLines / (linesPerItem + 1)) * perRow
		if itemsPerPage < perRow {
			itemsPerPage = perRow
		}
		return itemsPerPage
	}
	linesPerItem := 3
	itemsPerPage := availLines / (linesPerItem + 1)
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	return itemsPerPage
}

func renderQuizBottomSection(card ui.CardData, width, height, examplesPage int, topBodyLines int) string {
	sidePad := lipgloss.NewStyle().PaddingLeft(8).PaddingRight(8)
	var parts []string

	if card.Notes != "" {
		parts = append(parts, sidePad.Render(styles.MutedText().Render("note: "+card.Notes)))
	}

	if len(card.Tags) > 0 {
		parts = append(parts, sidePad.Render(renderQuizTags(card.Tags)))
	}

	if len(card.Examples) > 0 {
		parts = append(parts, sidePad.Render(renderQuizExamplesBlock(card.Examples, width, height, examplesPage, topBodyLines)))
	}

	return strings.Join(parts, "\n\n")
}
