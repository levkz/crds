package components

import "strings"

func ConfirmDialog(title, message, confirmLabel, cancelLabel string, width, height int) string {
	var b strings.Builder
	b.WriteString(message)
	b.WriteString("\n\n")
	b.WriteString("  " + confirmLabel + "  ")
	b.WriteString("  ")
	b.WriteString(cancelLabel)
	return RenderModal(title, b.String(), width, height)
}
