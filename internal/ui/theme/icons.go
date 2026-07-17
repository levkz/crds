package theme

type IconSource int

const (
	IconSourceAuto     IconSource = iota
	IconSourceNerdFont
	IconSourceEmoji
	IconSourceUnicode
	IconSourceFallback
)

func (s IconSource) String() string {
	switch s {
	case IconSourceNerdFont:
		return "nerdfont"
	case IconSourceEmoji:
		return "emoji"
	case IconSourceUnicode:
		return "unicode"
	case IconSourceFallback:
		return "fallback"
	default:
		return "auto"
	}
}

type Icons struct {
	Check     string
	Cross     string
	ArrowUp   string
	ArrowDown string
	ArrowLeft string
	Bullet    string

	Selected  string
	Navigate  string
	Highlight string
	Close     string
}

var NerdFontIcons = Icons{
	Check:     "",
	Cross:     "",
	ArrowUp:   "",
	ArrowDown: "",
	ArrowLeft: "",
	Bullet:    "",
	Selected:  "",
	Navigate:  "",
	Highlight: "",
	Close:     "",
}

var EmojiIcons = Icons{
	Check:     "✅",
	Cross:     "❌",
	ArrowUp:   "⬆",
	ArrowDown: "⬇",
	ArrowLeft: "⬅",
	Bullet:    "⭕",
	Selected:  "⭕",
	Navigate:  "➡",
	Highlight: "⭐",
	Close:     "❌",
}

var UnicodeIcons = Icons{
	Check:     "✓",
	Cross:     "✗",
	ArrowUp:   "▲",
	ArrowDown: "▼",
	ArrowLeft: "◀",
	Bullet:    "•",
	Selected:  "•",
	Navigate:  "▶",
	Highlight: "★",
	Close:     "✗",
}

var FallbackIcons = Icons{
	Check:     "[x]",
	Cross:     "[ ]",
	ArrowUp:   "^",
	ArrowDown: "v",
	ArrowLeft: "<",
	Bullet:    "*",
	Selected:  "*",
	Navigate:  ">",
	Highlight: "*",
	Close:     "[ ]",
}

func (i Icons) Fallback() Icons {
	return FallbackIcons
}

var DefaultIcons = UnicodeIcons

func IconsFromSource(s IconSource) Icons {
	switch s {
	case IconSourceNerdFont:
		return NerdFontIcons
	case IconSourceEmoji:
		return EmojiIcons
	case IconSourceUnicode:
		return UnicodeIcons
	case IconSourceFallback:
		return FallbackIcons
	default:
		return UnicodeIcons
	}
}
