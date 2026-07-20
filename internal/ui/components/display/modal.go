package components

import "crds/internal/ui/styles"

func RenderModal(title, content string, width, height int) string {
	s := styles.Modal(width, height)
	return s.Render(title + "\n\n" + content)
}
