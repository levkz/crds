package theme

import (
	"os"
	"strings"
)

func UnicodeSupported() bool {
	return lookupUTF8("LC_ALL") || lookupUTF8("LC_CTYPE") || lookupUTF8("LANG")
}

func lookupUTF8(key string) bool {
	v := os.Getenv(key)
	if v == "" {
		return false
	}
	u := strings.ToUpper(v)
	return strings.Contains(u, "UTF-8") || strings.Contains(u, "UTF8")
}

func EmojiSupported() bool {
	if !UnicodeSupported() {
		return false
	}
	// Most modern terminals with Unicode support also support emoji.
	// Check COLORTERM as a proxy for terminal modernity.
	ct := os.Getenv("COLORTERM")
	if ct == "" {
		return false
	}
	return strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit")
}

func DetectIconSource() IconSource {
	if v := os.Getenv("CRDS_ICON_SOURCE"); v != "" {
		switch strings.ToLower(v) {
		case "nerdfont", "nerd":
			return IconSourceNerdFont
		case "emoji":
			return IconSourceEmoji
		case "unicode":
			return IconSourceUnicode
		case "fallback", "ascii":
			return IconSourceFallback
		}
	}
	if NerdFontSupported() {
		return IconSourceNerdFont
	}
	if EmojiSupported() {
		return IconSourceEmoji
	}
	if UnicodeSupported() {
		return IconSourceUnicode
	}
	return IconSourceFallback
}

func DetectedIcons() Icons {
	return IconsFromSource(DetectIconSource())
}
