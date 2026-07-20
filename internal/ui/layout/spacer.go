package layout

import "strings"

func VSpace(n int) string {
	if n < 0 {
		return ""
	}
	return strings.Repeat("\n", n)
}

func HSpace(n int) string {
	if n < 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}
