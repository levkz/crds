package components

import (
	"crds/internal/ui"
	"crds/internal/ui/styles"
)

type BadgeVariant int

const (
	BadgeSuccess BadgeVariant = iota
	BadgeError
	BadgeWarning
	BadgeInfo
)

func Badge(text string, variant BadgeVariant) string {
	icon := ""
	style := styles.MutedText()

	switch variant {
	case BadgeSuccess:
		icon = ui.Theme.Icons.Check + " "
		style = styles.Success()
	case BadgeError:
		icon = ui.Theme.Icons.Cross + " "
		style = styles.Error()
	case BadgeWarning:
		icon = "! "
		style = styles.Warning()
	case BadgeInfo:
		icon = ui.Theme.Icons.Bullet + " "
		style = styles.Hint()
	}

	return style.Render(icon + text)
}
