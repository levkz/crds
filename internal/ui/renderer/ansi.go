package renderer

import "strings"

func StripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			i++
			if i < len(s) {
				if s[i] == '[' {
					i++
					for i < len(s) && (s[i] < 'A' || s[i] > 'Z') && (s[i] < 'a' || s[i] > 'z') {
						i++
					}
				} else if s[i] >= 0x40 && s[i] <= 0x5F {
					// two-char C1 control code, consumed by i++ at top
				} else {
					b.WriteByte(s[i])
				}
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func CountANSISequences(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			i++
			if i < len(s) && s[i] == '[' {
				i++
				for i < len(s) && (s[i] < 'A' || s[i] > 'Z') && (s[i] < 'a' || s[i] > 'z') {
					i++
				}
				n++
			} else if i < len(s) && s[i] >= 0x40 && s[i] <= 0x5F {
				n++
			}
		}
	}
	return n
}
