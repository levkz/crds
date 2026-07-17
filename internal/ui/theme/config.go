package theme

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	yaml "go.yaml.in/yaml/v3"
)

type Config struct {
	Palette    *ConfigPalette    `yaml:"palette"`
	Icons      string            `yaml:"icons"`
	Typography *ConfigTypography `yaml:"typography"`
}

type ConfigPalette struct {
	Blue       string `yaml:"blue"`
	Green      string `yaml:"green"`
	Orange     string `yaml:"orange"`
	Red        string `yaml:"red"`
	Gray       string `yaml:"gray"`
	White      string `yaml:"white"`
	Background string `yaml:"background"`
	Selection  string `yaml:"selection"`
	Border     string `yaml:"border"`
	Link       string `yaml:"link"`
	Surface    string `yaml:"surface"`
	Primary    string `yaml:"primary"`
	Secondary  string `yaml:"secondary"`
	Accent     string `yaml:"accent"`
}

type ConfigTypography struct {
	Title    *ConfigTextRole `yaml:"title"`
	Subtitle *ConfigTextRole `yaml:"subtitle"`
	Body     *ConfigTextRole `yaml:"body"`
	Caption  *ConfigTextRole `yaml:"caption"`
	Emphasis *ConfigTextRole `yaml:"emphasis"`
	Key      *ConfigTextRole `yaml:"key"`
}

type ConfigTextRole struct {
	Color  string `yaml:"color"`
	Bold   *bool  `yaml:"bold"`
	Italic *bool  `yaml:"italic"`
}

func (c Config) Build() (Theme, error) {
	p := DefaultPalette
	if c.Palette != nil {
		if c.Palette.Blue != "" {
			p.Blue = mustColor(c.Palette.Blue)
		}
		if c.Palette.Green != "" {
			p.Green = mustColor(c.Palette.Green)
		}
		if c.Palette.Orange != "" {
			p.Orange = mustColor(c.Palette.Orange)
		}
		if c.Palette.Red != "" {
			p.Red = mustColor(c.Palette.Red)
		}
		if c.Palette.Gray != "" {
			p.Gray = mustColor(c.Palette.Gray)
		}
		if c.Palette.White != "" {
			p.White = mustColor(c.Palette.White)
		}
		if c.Palette.Background != "" {
			p.Background = mustColor(c.Palette.Background)
		}
		if c.Palette.Selection != "" {
			p.Selection = mustColor(c.Palette.Selection)
		}
		if c.Palette.Border != "" {
			p.Border = mustColor(c.Palette.Border)
		}
		if c.Palette.Link != "" {
			p.Link = mustColor(c.Palette.Link)
		}
		if c.Palette.Surface != "" {
			p.Surface = mustColor(c.Palette.Surface)
		}
	}

	th := NewTheme(p)

	if c.Palette != nil {
		if c.Palette.Primary != "" {
			color, err := paletteColor(p, c.Palette.Primary)
			if err != nil {
				return Theme{}, fmt.Errorf("palette.primary: %w", err)
			}
			th.Primary = th.Primary.Foreground(color)
		}
		if c.Palette.Secondary != "" {
			color, err := paletteColor(p, c.Palette.Secondary)
			if err != nil {
				return Theme{}, fmt.Errorf("palette.secondary: %w", err)
			}
			th.Secondary = th.Secondary.Foreground(color)
		}
		if c.Palette.Accent != "" {
			color, err := paletteColor(p, c.Palette.Accent)
			if err != nil {
				return Theme{}, fmt.Errorf("palette.accent: %w", err)
			}
			th.Accent = th.Accent.Foreground(color)
		}
	}

	if c.Icons != "" {
		switch c.Icons {
		case "nerdfont", "nerd":
			th.Icons = NerdFontIcons
		case "emoji":
			th.Icons = EmojiIcons
		case "unicode":
			th.Icons = UnicodeIcons
		case "fallback", "ascii":
			th.Icons = FallbackIcons
		default:
			return Theme{}, fmt.Errorf("unknown icon source %q", c.Icons)
		}
	}

	if c.Typography != nil {
		var err error
		th.Typography.Title, err = applyTextRole(th.Typography.Title, c.Typography.Title, p)
		if err != nil {
			return Theme{}, fmt.Errorf("typography.title: %w", err)
		}
		th.Typography.Subtitle, err = applyTextRole(th.Typography.Subtitle, c.Typography.Subtitle, p)
		if err != nil {
			return Theme{}, fmt.Errorf("typography.subtitle: %w", err)
		}
		th.Typography.Body, err = applyTextRole(th.Typography.Body, c.Typography.Body, p)
		if err != nil {
			return Theme{}, fmt.Errorf("typography.body: %w", err)
		}
		th.Typography.Caption, err = applyTextRole(th.Typography.Caption, c.Typography.Caption, p)
		if err != nil {
			return Theme{}, fmt.Errorf("typography.caption: %w", err)
		}
		th.Typography.Emphasis, err = applyTextRole(th.Typography.Emphasis, c.Typography.Emphasis, p)
		if err != nil {
			return Theme{}, fmt.Errorf("typography.emphasis: %w", err)
		}
		th.Typography.Key, err = applyTextRole(th.Typography.Key, c.Typography.Key, p)
		if err != nil {
			return Theme{}, fmt.Errorf("typography.key: %w", err)
		}
	}

	return th, nil
}

func applyTextRole(base lipgloss.Style, cfg *ConfigTextRole, p Palette) (lipgloss.Style, error) {
	if cfg == nil {
		return base, nil
	}
	s := base
	if cfg.Color != "" {
		c, err := paletteColor(p, cfg.Color)
		if err != nil {
			return s, err
		}
		s = s.Foreground(c)
	}
	if cfg.Bold != nil {
		s = s.Bold(*cfg.Bold)
	}
	if cfg.Italic != nil {
		s = s.Italic(*cfg.Italic)
	}
	return s, nil
}

func paletteColor(p Palette, name string) (lipgloss.Color, error) {
	switch strings.ToLower(name) {
	case "blue":
		return p.Blue, nil
	case "green":
		return p.Green, nil
	case "orange":
		return p.Orange, nil
	case "red":
		return p.Red, nil
	case "gray":
		return p.Gray, nil
	case "white":
		return p.White, nil
	case "background":
		return p.Background, nil
	case "selection":
		return p.Selection, nil
	case "border":
		return p.Border, nil
	case "link":
		return p.Link, nil
	case "surface":
		return p.Surface, nil
	default:
		return resolveDirectColor(name)
	}
}

func resolveDirectColor(s string) (lipgloss.Color, error) {
	if s == "" {
		return "", fmt.Errorf("empty color value")
	}
	if s[0] == '#' || (s[0] >= '0' && s[0] <= '9') {
		return lipgloss.Color(s), nil
	}
	return "", fmt.Errorf("unknown color %q", s)
}

func mustColor(s string) lipgloss.Color {
	if s == "" {
		panic("color must not be empty")
	}
	return lipgloss.Color(s)
}

func LoadTheme(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, fmt.Errorf("reading theme file: %w", err)
	}
	return ParseTheme(data)
}

func ParseTheme(data []byte) (Theme, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Theme{}, fmt.Errorf("parsing theme config: %w", err)
	}
	return cfg.Build()
}
