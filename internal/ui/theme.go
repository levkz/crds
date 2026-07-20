package ui

import "crds/internal/ui/theme"

var Theme = theme.Default

func SetTheme(t theme.Theme) {
	Theme = t
}
