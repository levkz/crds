package components

import "crds/internal/ui/styles"

func ErrorDialog(title, message string, width, height int) string {
	return RenderModal(title, styles.Error().Render(message), width, height)
}
