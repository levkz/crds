package components

import (
	"fmt"

	"crds/internal/ui/styles"
)

func ProgressBar(progress int) string {
	return styles.MutedText().Render(
		fmt.Sprintf("Progress: %d%%", progress),
	)
}
