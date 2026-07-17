package components

import "crds/internal/ui/styles"

func RenderModal(title, content string) string {
	s := styles.Modal(40, 10)
	return s.Render(title + "\n\n" + content)
}
