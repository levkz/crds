package layout

import "strings"

func Column(items ...string) string {
	return strings.Join(items, "\n\n")
}
