package theme

import (
	"os"
	"strings"
)

func NerdFontSupported() bool {
	if v := os.Getenv("CRDS_NERD_FONT"); v == "1" || strings.EqualFold(v, "true") {
		return true
	}
	if !UnicodeSupported() {
		return false
	}
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}
	// NerdFont-compatible terminal types:
	//   xterm-256color, xterm-kitty, xterm, tmux-256color,
	//   screen-256color, alacritty, wezterm, foot,
	//   vscode, ghostty, iTerm.app, st-256color
	nerdTerms := []string{
		"xterm-256color", "xterm-kitty", "xterm",
		"tmux-256color", "screen-256color",
		"alacritty", "wezterm", "foot", "foot-extra",
		"contour", "rio",
	}
	for _, nt := range nerdTerms {
		if strings.EqualFold(term, nt) {
			return true
		}
	}
	if strings.HasPrefix(term, "xterm") {
		return true
	}
	return false
}
