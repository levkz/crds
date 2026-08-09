package components

import (
	"fmt"
	"strings"

	"crds/internal/stats"
	"crds/internal/ui"
	"crds/internal/ui/renderer"
)

// GraphPoint is one day of review outcomes for the graph.
type GraphPoint struct {
	Day       string // short label, e.g. "04-08"
	Correct   int
	Incorrect int
}

// Graph renders a per-day confidence bar chart:
//
//	04-08 █████████░ 90% (9/1)
//
// Bars are scaled to confidence (correct/total); days with no reviews show a
// neutral marker. The bar color reflects confidence: high (>=70%) success,
// mid (>=40%) warning, low danger.
func Graph(points []GraphPoint, width int) string {
	if len(points) == 0 {
		return ui.Theme.Muted.Render("No reviews yet")
	}
	if width < 1 {
		return ""
	}

	// Reserve: day label (6) + spacing + " NN% (N/N)" suffix.
	// Compute the widest suffix to keep bars aligned.
	maxDayW := 0
	for _, p := range points {
		if w := renderer.VisibleWidth(p.Day); w > maxDayW {
			maxDayW = w
		}
	}

	suffix := " 100% (999/999)"
	suffixW := renderer.VisibleWidth(suffix)
	barW := width - maxDayW - suffixW
	if barW < 3 {
		barW = 3
	}

	var b strings.Builder
	for i, p := range points {
		if i > 0 {
			b.WriteString("\n")
		}
		day := p.Day + strings.Repeat(" ", maxDayW-renderer.VisibleWidth(p.Day))
		b.WriteString(ui.Theme.Muted.Render(day))

		total := p.Correct + p.Incorrect
		var bar string
		if total == 0 {
			bar = renderer.Truncate(strings.Repeat("─", barW), barW)
			b.WriteString(" " + ui.Theme.Muted.Render(bar))
		} else {
			conf := float64(p.Correct) / float64(total)
			filled := int(float64(barW) * conf)
			if filled > barW {
				filled = barW
			}
			style := ui.Theme.Success
			switch {
			case conf < 0.4:
				style = ui.Theme.Danger
			case conf < 0.7:
				style = ui.Theme.Warning
			}
			bar = style.Render(strings.Repeat("█", filled))
			if filled < barW {
				bar += ui.Theme.Muted.Render(strings.Repeat("░", barW-filled))
			}
			b.WriteString(" " + bar)
		}

		pct := ""
		if total > 0 {
			pct = fmt.Sprintf(" %d%%", int(float64(p.Correct)/float64(total)*100))
		}
		suffix := pct + fmt.Sprintf(" (%d/%d)", p.Correct, total)
		b.WriteString(ui.Theme.Muted.Render(suffix))
	}
	return b.String()
}

// ToGraphPoints converts stats.DayPoint values (full YYYY-MM-DD) into
// GraphPoints with MM-DD labels.
func ToGraphPoints(days []stats.DayPoint) []GraphPoint {
	out := make([]GraphPoint, 0, len(days))
	for _, d := range days {
		label := d.Day
		if len(d.Day) >= 10 {
			label = d.Day[5:]
		}
		out = append(out, GraphPoint{
			Day:       label,
			Correct:   d.Correct,
			Incorrect: d.Incorrect,
		})
	}
	return out
}
