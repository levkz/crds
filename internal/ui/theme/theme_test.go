package theme

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultPalette(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
		want  string
	}{
		{"Blue", DefaultPalette.Blue, "39"},
		{"Green", DefaultPalette.Green, "42"},
		{"Orange", DefaultPalette.Orange, "214"},
		{"Red", DefaultPalette.Red, "196"},
		{"Gray", DefaultPalette.Gray, "248"},
		{"White", DefaultPalette.White, "255"},
		{"Background", DefaultPalette.Background, "0"},
		{"Selection", DefaultPalette.Selection, "27"},
		{"Border", DefaultPalette.Border, "59"},
		{"Link", DefaultPalette.Link, "33"},
		{"Surface", DefaultPalette.Surface, "235"},
		{"Magenta", DefaultPalette.Magenta, "177"},
		{"Purple", DefaultPalette.Purple, "140"},
		{"Cyan", DefaultPalette.Cyan, "117"},
		{"Yellow", DefaultPalette.Yellow, "220"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.color); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	th := Default
	if th.Palette != DefaultPalette {
		t.Error("Default theme should use DefaultPalette")
	}
	if th.Icons.Check != DefaultIcons.Check {
		t.Error("Default theme should use DefaultIcons")
	}
}

func TestThemeStylesRender(t *testing.T) {
	th := Default
	styles := []struct {
		name string
		ns   lipgloss.Style
	}{
		{"Primary", th.Primary},
		{"Secondary", th.Secondary},
		{"Accent", th.Accent},
		{"Success", th.Success},
		{"Warning", th.Warning},
		{"Danger", th.Danger},
		{"Muted", th.Muted},
		{"Header", th.Header},
		{"Background", th.Background},
		{"Surface", th.Surface},
	}
	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			if s.ns.Render("x") == "" {
				t.Errorf("%s.Render returned empty", s.name)
			}
		})
	}
}

func TestNewThemeCustomPalette(t *testing.T) {
	p := Palette{
		Blue:       lipgloss.Color("1"),
		Green:      lipgloss.Color("2"),
		Orange:     lipgloss.Color("3"),
		Red:        lipgloss.Color("4"),
		Gray:       lipgloss.Color("5"),
		White:      lipgloss.Color("6"),
		Background: lipgloss.Color("7"),
		Selection:  lipgloss.Color("8"),
		Border:     lipgloss.Color("9"),
		Link:       lipgloss.Color("10"),
		Surface:    lipgloss.Color("11"),
	}
	th := NewTheme(p)
	if string(th.Palette.Blue) != "1" {
		t.Error("custom palette not applied")
	}
	rendered := th.Primary.Render("x")
	if rendered == "" {
		t.Error("Primary.Render returned empty with custom palette")
	}
}

func TestIconsDefaults(t *testing.T) {
	if DefaultIcons.Check != "✓" {
		t.Error("expected check icon")
	}
}

func TestIconsFallback(t *testing.T) {
	fb := DefaultIcons.Fallback()
	if fb.Check != "[x]" {
		t.Errorf("expected '[x]', got %q", fb.Check)
	}
	if fb.ArrowUp != "^" {
		t.Errorf("expected '^', got %q", fb.ArrowUp)
	}
}

func TestTypographyStyles(t *testing.T) {
	th := Default
	styles := []struct {
		name string
		s    lipgloss.Style
	}{
		{"Title", th.Typography.Title},
		{"Subtitle", th.Typography.Subtitle},
		{"Body", th.Typography.Body},
		{"Caption", th.Typography.Caption},
		{"Emphasis", th.Typography.Emphasis},
		{"Key", th.Typography.Key},
	}
	for _, st := range styles {
		t.Run(st.name, func(t *testing.T) {
			if st.s.Render("x") == "" {
				t.Errorf("%s.Render returned empty", st.name)
			}
		})
	}
}

func TestBorders(t *testing.T) {
	b := DefaultBorders

	tests := []struct {
		name string
		bdr  lipgloss.Border
	}{
		{"Normal", b.Normal},
		{"Rounded", b.Rounded},
		{"Double", b.Double},
		{"Thick", b.Thick},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bdr.Top == "" || tt.bdr.Bottom == "" ||
				tt.bdr.Left == "" || tt.bdr.Right == "" {
				t.Errorf("%s border has empty edge characters", tt.name)
			}
		})
	}
}

func TestBordersApply(t *testing.T) {
	style := lipgloss.NewStyle().Border(DefaultBorders.Rounded)
	rendered := style.Render("x")
	if rendered == "" {
		t.Error("Rendered bordered style is empty")
	}
}

func TestNewThemeBorders(t *testing.T) {
	th := NewTheme(DefaultPalette)
	if th.Borders.Normal.Top != DefaultBorders.Normal.Top {
		t.Error("NewTheme should use DefaultBorders")
	}
}

func TestNewThemeTypography(t *testing.T) {
	th := NewTheme(DefaultPalette)
	rendered := th.Typography.Title.Render("Test Title")
	if rendered == "" {
		t.Error("Title.Render returned empty")
	}
}

func TestUnicodeSupportedEnv(t *testing.T) {
	t.Run("LC_ALL=UTF-8", func(t *testing.T) {
		t.Setenv("LC_ALL", "en_US.UTF-8")
		if !UnicodeSupported() {
			t.Error("expected Unicode supported with LC_ALL=UTF-8")
		}
	})
	t.Run("LC_CTYPE=C", func(t *testing.T) {
		t.Setenv("LC_ALL", "")
		t.Setenv("LC_CTYPE", "C")
		t.Setenv("LANG", "C")
		if UnicodeSupported() {
			t.Error("expected no Unicode support with LC_CTYPE=C")
		}
	})
	t.Run("no UTF-8 env", func(t *testing.T) {
		t.Setenv("LC_ALL", "")
		t.Setenv("LC_CTYPE", "")
		t.Setenv("LANG", "C")
		if UnicodeSupported() {
			t.Error("expected no Unicode support without UTF-8 in env")
		}
	})
}

func TestDetectedIcons(t *testing.T) {
	t.Run("unicode supported", func(t *testing.T) {
		t.Setenv("TERM", "dumb")
		t.Setenv("COLORTERM", "")
		t.Setenv("LANG", "en_US.UTF-8")
		icons := DetectedIcons()
		if icons.Check != "✓" {
			t.Errorf("expected Unicode checkmark, got %q", icons.Check)
		}
	})
	t.Run("no unicode", func(t *testing.T) {
		t.Setenv("TERM", "dumb")
		t.Setenv("COLORTERM", "")
		t.Setenv("LANG", "C")
		icons := DetectedIcons()
		if icons.Check != "[x]" {
			t.Errorf("expected fallback [x], got %q", icons.Check)
		}
	})
}

func TestNerdFontDetection(t *testing.T) {
	tests := []struct {
		name string
		term string
		want bool
	}{
		{"xterm-256color", "xterm-256color", true},
		{"kitty", "xterm-kitty", true},
		{"alacritty", "alacritty", true},
		{"wezterm", "wezterm", true},
		{"foot", "foot", true},
		{"tmux-256color", "tmux-256color", true},
		{"st-256color", "st-256color", false},
		{"linux", "linux", false},
		{"dumb", "dumb", false},
		{"empty", "", false},
		{"CRDS_NERD_FONT=1", "dumb", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			t.Setenv("LANG", "en_US.UTF-8")
			if tt.name == "CRDS_NERD_FONT=1" {
				t.Setenv("CRDS_NERD_FONT", "1")
				t.Setenv("LANG", "C")
			} else {
				t.Setenv("CRDS_NERD_FONT", "")
			}
			if got := NerdFontSupported(); got != tt.want {
				t.Errorf("NerdFontSupported() = %v, want %v (TERM=%q)", got, tt.want, tt.term)
			}
		})
	}
}

func TestNerdFontIcons(t *testing.T) {
	if NerdFontIcons.Check == "" {
		t.Error("NerdFontIcons.Check is empty")
	}
	if NerdFontIcons.ArrowUp == NerdFontIcons.ArrowDown {
		t.Error("ArrowUp and ArrowDown should differ")
	}
}

func TestEmojiDetection(t *testing.T) {
	t.Run("truecolor terminal", func(t *testing.T) {
		t.Setenv("COLORTERM", "truecolor")
		t.Setenv("LANG", "en_US.UTF-8")
		if !EmojiSupported() {
			t.Error("expected emoji supported with COLORTERM=truecolor")
		}
	})
	t.Run("no truecolor", func(t *testing.T) {
		t.Setenv("COLORTERM", "")
		t.Setenv("LANG", "en_US.UTF-8")
		if EmojiSupported() {
			t.Error("expected no emoji without COLORTERM")
		}
	})
	t.Run("no unicode", func(t *testing.T) {
		t.Setenv("COLORTERM", "truecolor")
		t.Setenv("LANG", "C")
		if EmojiSupported() {
			t.Error("expected no emoji without Unicode")
		}
	})
}

func TestEmojiIcons(t *testing.T) {
	if EmojiIcons.Check == "" {
		t.Error("EmojiIcons.Check is empty")
	}
}

func TestSemanticIconDefaults(t *testing.T) {
	if DefaultIcons.Selected == "" {
		t.Error("DefaultIcons.Selected is empty")
	}
	if DefaultIcons.Navigate == "" {
		t.Error("DefaultIcons.Navigate is empty")
	}
	if DefaultIcons.Highlight == "" {
		t.Error("DefaultIcons.Highlight is empty")
	}
	if DefaultIcons.Close == "" {
		t.Error("DefaultIcons.Close is empty")
	}
}

func TestSemanticIconsAllSources(t *testing.T) {
	sources := []struct {
		name string
		icons Icons
	}{
		{"NerdFont", NerdFontIcons},
		{"Emoji", EmojiIcons},
		{"Unicode", UnicodeIcons},
		{"Fallback", FallbackIcons},
	}
	for _, s := range sources {
		t.Run(s.name, func(t *testing.T) {
			if s.icons.Selected == "" {
				t.Errorf("%s.Selected is empty", s.name)
			}
			if s.icons.Navigate == "" {
				t.Errorf("%s.Navigate is empty", s.name)
			}
			if s.icons.Highlight == "" {
				t.Errorf("%s.Highlight is empty", s.name)
			}
			if s.icons.Close == "" {
				t.Errorf("%s.Close is empty", s.name)
			}
		})
	}
}

func TestDetectIconSource(t *testing.T) {
	t.Run("nerdfont preferred", func(t *testing.T) {
		t.Setenv("TERM", "alacritty")
		t.Setenv("LANG", "en_US.UTF-8")
		if src := DetectIconSource(); src != IconSourceNerdFont {
			t.Errorf("expected NerdFont, got %s", src)
		}
	})
	t.Run("emoji when no nerdfont", func(t *testing.T) {
		t.Setenv("TERM", "linux")
		t.Setenv("COLORTERM", "truecolor")
		t.Setenv("LANG", "en_US.UTF-8")
		if src := DetectIconSource(); src != IconSourceEmoji {
			t.Errorf("expected Emoji, got %s", src)
		}
	})
	t.Run("unicode fallback", func(t *testing.T) {
		t.Setenv("TERM", "linux")
		t.Setenv("COLORTERM", "")
		t.Setenv("LANG", "en_US.UTF-8")
		if src := DetectIconSource(); src != IconSourceUnicode {
			t.Errorf("expected Unicode, got %s", src)
		}
	})
	t.Run("ascii fallback", func(t *testing.T) {
		t.Setenv("TERM", "linux")
		t.Setenv("COLORTERM", "")
		t.Setenv("LANG", "C")
		if src := DetectIconSource(); src != IconSourceFallback {
			t.Errorf("expected Fallback, got %s", src)
		}
	})
}

func TestDetectIconSourceOverride(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want IconSource
	}{
		{"nerdfont", "nerdfont", IconSourceNerdFont},
		{"nerd", "nerd", IconSourceNerdFont},
		{"emoji", "emoji", IconSourceEmoji},
		{"unicode", "unicode", IconSourceUnicode},
		{"fallback", "fallback", IconSourceFallback},
		{"ascii", "ascii", IconSourceFallback},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CRDS_ICON_SOURCE", tt.val)
			t.Setenv("TERM", "linux")
			t.Setenv("LANG", "C")
			if src := DetectIconSource(); src != tt.want {
				t.Errorf("got %s, want %s", src, tt.want)
			}
		})
	}
}

func TestIconsFromSource(t *testing.T) {
	tests := []struct {
		name string
		src  IconSource
		want string
	}{
		{"NerdFont", IconSourceNerdFont, ""},
		{"Emoji", IconSourceEmoji, "✅"},
		{"Unicode", IconSourceUnicode, "✓"},
		{"Fallback", IconSourceFallback, "[x]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icons := IconsFromSource(tt.src)
			if icons.Check != tt.want {
				t.Errorf("got %q, want %q", icons.Check, tt.want)
			}
		})
	}
}

func TestIconSourceString(t *testing.T) {
	tests := []struct {
		src  IconSource
		want string
	}{
		{IconSourceNerdFont, "nerdfont"},
		{IconSourceEmoji, "emoji"},
		{IconSourceUnicode, "unicode"},
		{IconSourceFallback, "fallback"},
		{IconSourceAuto, "auto"},
		{IconSource(99), "auto"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.src.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithFallbackIcons(t *testing.T) {
	th := Default.WithFallbackIcons()
	if th.Icons.Check != "[x]" {
		t.Errorf("expected [x], got %q", th.Icons.Check)
	}
}

func TestNewTerminalTheme(t *testing.T) {
	th := NewTerminalTheme()
	if th.Palette != DefaultPalette {
		t.Error("NewTerminalTheme should use DefaultPalette")
	}
	if th.Icons.Check == "" {
		t.Error("icons should be populated")
	}
}

func TestNewThemePreservesIcons(t *testing.T) {
	th := NewTheme(DefaultPalette)
	if th.Icons.Bullet != "•" {
		t.Error("NewTheme should use DefaultIcons")
	}
}

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("testdata", name)
}

func TestLoadThemeFull(t *testing.T) {
	th, err := LoadTheme(testdataPath(t, "full.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(th.Palette.Blue) != "21" {
		t.Errorf("Blue = %q, want 21", th.Palette.Blue)
	}
	if string(th.Palette.Green) != "46" {
		t.Errorf("Green = %q, want 46", th.Palette.Green)
	}
	if string(th.Palette.Orange) != "208" {
		t.Errorf("Orange = %q, want 208", th.Palette.Orange)
	}
	if string(th.Palette.Red) != "160" {
		t.Errorf("Red = %q, want 160", th.Palette.Red)
	}
	if string(th.Palette.Gray) != "242" {
		t.Errorf("Gray = %q, want 242", th.Palette.Gray)
	}
	if string(th.Palette.White) != "231" {
		t.Errorf("White = %q, want 231", th.Palette.White)
	}
	if string(th.Palette.Background) != "235" {
		t.Errorf("Background = %q, want 235", th.Palette.Background)
	}
	if string(th.Palette.Selection) != "25" {
		t.Errorf("Selection = %q, want 25", th.Palette.Selection)
	}
	if string(th.Palette.Border) != "237" {
		t.Errorf("Border = %q, want 237", th.Palette.Border)
	}
	if string(th.Palette.Link) != "33" {
		t.Errorf("Link = %q, want 33", th.Palette.Link)
	}
	if string(th.Palette.Surface) != "236" {
		t.Errorf("Surface = %q, want 236", th.Palette.Surface)
	}
	if string(th.Palette.Magenta) != "177" {
		t.Errorf("Magenta = %q, want 177", th.Palette.Magenta)
	}
	if string(th.Palette.Purple) != "140" {
		t.Errorf("Purple = %q, want 140", th.Palette.Purple)
	}
	if string(th.Palette.Cyan) != "117" {
		t.Errorf("Cyan = %q, want 117", th.Palette.Cyan)
	}
	if string(th.Palette.Yellow) != "220" {
		t.Errorf("Yellow = %q, want 220", th.Palette.Yellow)
	}
	fs := th.Primary.GetForeground()
	if fs != th.Palette.Green {
		t.Errorf("Primary foreground = %v, want Green=%v", fs, th.Palette.Green)
	}
	ss := th.Secondary.GetForeground()
	if ss != th.Palette.Orange {
		t.Errorf("Secondary foreground = %v, want Orange=%v", ss, th.Palette.Orange)
	}
	as := th.Accent.GetForeground()
	if as != th.Palette.Red {
		t.Errorf("Accent foreground = %v, want Red=%v", as, th.Palette.Red)
	}
	if th.Icons.Check != "" {
		t.Errorf("Icons = %q, want NerdFont check", th.Icons.Check)
	}
	if th.Typography.Title.Render("x") == "" {
		t.Error("Title.Render empty after loading full theme")
	}
}

func TestLoadThemePartial(t *testing.T) {
	th, err := LoadTheme(testdataPath(t, "partial.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(th.Palette.Blue) != "21" {
		t.Errorf("Blue = %q, want 21", th.Palette.Blue)
	}
	// unset fields should default
	if string(th.Palette.Red) != "196" {
		t.Errorf("Red = %q, want default 196", th.Palette.Red)
	}
	if th.Icons.Check != "✓" {
		t.Errorf("Icons = %q, want default Unicode check", th.Icons.Check)
	}
}

func TestLoadThemeIconsOnly(t *testing.T) {
	th, err := LoadTheme(testdataPath(t, "icons_only.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if th.Icons.Check != "✅" {
		t.Errorf("Icons = %q, want emoji check", th.Icons.Check)
	}
	// palette should default
	if string(th.Palette.Blue) != "39" {
		t.Errorf("Blue = %q, want default 39", th.Palette.Blue)
	}
}

func TestLoadThemeMissingFile(t *testing.T) {
	_, err := LoadTheme("testdata/nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadThemeTypography(t *testing.T) {
	th, err := LoadTheme(testdataPath(t, "typography.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if th.Typography.Title.Render("x") == "" {
		t.Error("Title.Render empty after typography config")
	}
	// palette should still be defaults
	if string(th.Palette.Blue) != "39" {
		t.Errorf("Blue = %q, want default 39", th.Palette.Blue)
	}
}

func TestParseThemeTypography(t *testing.T) {
	data := []byte(`typography:
  title:
    color: green
    bold: false
    italic: true`)
	th, err := ParseTheme(data)
	if err != nil {
		t.Fatal(err)
	}
	if th.Typography.Title.Render("x") == "" {
		t.Error("Title.Render empty after typography parse")
	}
}

func TestParseThemeTypographyUnknownColor(t *testing.T) {
	data := []byte(`typography:
  title:
    color: nonexistent`)
	_, err := ParseTheme(data)
	if err == nil {
		t.Fatal("expected error for unknown color")
	}
}

func TestLoadThemeInvalidIcons(t *testing.T) {
	_, err := LoadTheme(testdataPath(t, "invalid_icons.yaml"))
	if err == nil {
		t.Fatal("expected error for invalid icon source")
	}
}

func TestLoadThemeMalformed(t *testing.T) {
	_, err := LoadTheme(testdataPath(t, "malformed.yaml"))
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestParseThemeFull(t *testing.T) {
	data := []byte(`palette:
  blue: "21"
  green: "46"
icons: emoji`)
	th, err := ParseTheme(data)
	if err != nil {
		t.Fatal(err)
	}
	if string(th.Palette.Blue) != "21" {
		t.Errorf("Blue = %q, want 21", th.Palette.Blue)
	}
	if th.Icons.Check != "✅" {
		t.Errorf("Icons = %q, want emoji", th.Icons.Check)
	}
}

func TestParseThemeEmpty(t *testing.T) {
	th, err := ParseTheme([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if string(th.Palette.Blue) != "39" {
		t.Errorf("Blue = %q, want default 39", th.Palette.Blue)
	}
	if th.Icons.Check != "✓" {
		t.Errorf("Icons = %q, want default check", th.Icons.Check)
	}
}

func TestConfigBuildMissingFile(t *testing.T) {
	// Verify error path when os.ReadFile fails
	_, err := LoadTheme(testdataPath(t, "not_a_file.yaml"))
	if err == nil {
		t.Error("expected error")
	}
}

func TestLoadThemeCreatesStyles(t *testing.T) {
	th, err := LoadTheme(testdataPath(t, "full.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if th.Primary.Render("x") == "" {
		t.Error("Primary.Render empty after loading")
	}
	if th.Typography.Title.Render("x") == "" {
		t.Error("Typography.Title.Render empty after loading")
	}
	if th.Borders.Rounded.Top == "" {
		t.Error("Borders.Rounded empty after loading")
	}
}

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s.Len() != 5 {
		t.Errorf("expected 5 themes (default, dark, light, tokyonight, mocha), got %d", s.Len())
	}
	if name := s.CurrentName(); name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}
	if s.Current().Palette != DefaultPalette {
		t.Error("current theme should be Default")
	}
	if !s.Has("dark") {
		t.Error("expected 'dark' to be registered")
	}
	if !s.Has("light") {
		t.Error("expected 'light' to be registered")
	}
}

func TestStoreRegister(t *testing.T) {
	s := NewStore()
	s.Register("custom", NewTheme(DefaultPalette))
	if !s.Has("custom") {
		t.Error("expected 'custom' to be registered")
	}
	if s.Len() != 6 {
		t.Errorf("expected 6 themes, got %d", s.Len())
	}
}

func TestStoreSwitch(t *testing.T) {
	s := NewStore()

	p := Palette{Blue: "1", Green: "2", Orange: "3", Red: "4", Gray: "5", White: "6", Background: "0", Selection: "27", Border: "59", Link: "33", Surface: "11"}
	s.Register("custom", NewTheme(p))

	th, err := s.Switch("custom")
	if err != nil {
		t.Fatal(err)
	}
	if s.CurrentName() != "custom" {
		t.Errorf("expected 'custom', got %q", s.CurrentName())
	}
	if string(th.Palette.Blue) != "1" {
		t.Errorf("expected Blue=1, got %q", th.Palette.Blue)
	}
	// verify Current() also returns the switched theme
	if string(s.Current().Palette.Blue) != "1" {
		t.Error("Current() should return switched theme")
	}
}

func TestStoreSwitchUnknown(t *testing.T) {
	s := NewStore()
	_, err := s.Switch("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown theme")
	}
}

func TestStoreStartsWithDefault(t *testing.T) {
	s := NewStore()
	if !s.Has("default") {
		t.Error("Store should start with 'default' registered")
	}
}

func TestStoreNames(t *testing.T) {
	s := NewStore()
	s.Register("a", NewTheme(DefaultPalette))
	s.Register("b", NewTheme(DefaultPalette))
	names := s.Names()
	sort.Strings(names)
	expected := []string{"a", "b", "dark", "default", "light", "mocha", "tokyonight"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d: %v", len(expected), len(names), names)
	}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("names[%d] = %q, want %q", i, n, expected[i])
		}
	}
}

func TestStoreGet(t *testing.T) {
	s := NewStore()
	s.Register("custom", NewTheme(DefaultPalette))
	th, ok := s.Get("custom")
	if !ok {
		t.Fatal("expected 'custom' to exist")
	}
	if th.Palette != DefaultPalette {
		t.Error("expected DefaultPalette")
	}
	_, ok = s.Get("missing")
	if ok {
		t.Error("expected 'missing' to not exist")
	}
}

func TestStoreUnregister(t *testing.T) {
	s := NewStore()
	s.Register("temp", NewTheme(DefaultPalette))
	s.Switch("temp")
	s.Unregister("temp")
	if s.Has("temp") {
		t.Error("expected 'temp' to be unregistered")
	}
	if s.CurrentName() != "default" {
		t.Error("should fall back to 'default' after unregistering current")
	}
}

func TestDarkTheme(t *testing.T) {
	th := DarkTheme()
	if string(th.Palette.Blue) != "75" {
		t.Errorf("Dark Blue = %q, want 75", th.Palette.Blue)
	}
	if string(th.Palette.Background) != "233" {
		t.Errorf("Dark Background = %q, want 233", th.Palette.Background)
	}
	if string(th.Palette.Surface) != "236" {
		t.Errorf("Dark Surface = %q, want 236", th.Palette.Surface)
	}
	if th.Icons.Check == "" {
		t.Error("Dark theme icons should be populated")
	}
}

func TestLightTheme(t *testing.T) {
	th := LightTheme()
	if string(th.Palette.Blue) != "27" {
		t.Errorf("Light Blue = %q, want 27", th.Palette.Blue)
	}
	if string(th.Palette.White) != "235" {
		t.Errorf("Light White = %q, want 235", th.Palette.White)
	}
	if string(th.Palette.Background) != "255" {
		t.Errorf("Light Background = %q, want 255", th.Palette.Background)
	}
	if string(th.Palette.Surface) != "250" {
		t.Errorf("Light Surface = %q, want 250", th.Palette.Surface)
	}
	if th.Icons.Check == "" {
		t.Error("Light theme icons should be populated")
	}
}

func TestTokyonightPalette(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
		want  string
	}{
		{"Blue", TokyonightPalette.Blue, "#7aa2f7"},
		{"Green", TokyonightPalette.Green, "#9ece6a"},
		{"Orange", TokyonightPalette.Orange, "#ff9e64"},
		{"Red", TokyonightPalette.Red, "#f7768e"},
		{"Gray", TokyonightPalette.Gray, "#7982a8"},
		{"White", TokyonightPalette.White, "#c0caf5"},
		{"Background", TokyonightPalette.Background, "#1a1b26"},
		{"Selection", TokyonightPalette.Selection, "#283457"},
		{"Border", TokyonightPalette.Border, "#15161e"},
		{"Link", TokyonightPalette.Link, "#2ac3de"},
		{"Surface", TokyonightPalette.Surface, "#16161e"},
		{"Magenta", TokyonightPalette.Magenta, "#bb9af7"},
		{"Purple", TokyonightPalette.Purple, "#9d7cd8"},
		{"Cyan", TokyonightPalette.Cyan, "#7dcfff"},
		{"Yellow", TokyonightPalette.Yellow, "#e0af68"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.color); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokyonightTheme(t *testing.T) {
	th := TokyonightTheme()
	if string(th.Palette.Blue) != "#7aa2f7" {
		t.Errorf("Tokyonight Blue = %q, want #7aa2f7", th.Palette.Blue)
	}
	if string(th.Palette.Background) != "#1a1b26" {
		t.Errorf("Tokyonight Background = %q, want #1a1b26", th.Palette.Background)
	}
	if string(th.Palette.Surface) != "#16161e" {
		t.Errorf("Tokyonight Surface = %q, want #16161e", th.Palette.Surface)
	}
	if th.Icons.Check == "" {
		t.Error("Tokyonight theme icons should be populated")
	}
}

func TestDefaultSpacing(t *testing.T) {
	s := DefaultSpacing
	if s.Xxs != 2 {
		t.Errorf("Xxs = %d, want 2", s.Xxs)
	}
	if s.Xs != 4 {
		t.Errorf("Xs = %d, want 4", s.Xs)
	}
	if s.Sm != 8 {
		t.Errorf("Sm = %d, want 8", s.Sm)
	}
	if s.Md != 16 {
		t.Errorf("Md = %d, want 16", s.Md)
	}
	if s.Lg != 24 {
		t.Errorf("Lg = %d, want 24", s.Lg)
	}
	if s.Xl != 32 {
		t.Errorf("Xl = %d, want 32", s.Xl)
	}
	if s.Xxl != 48 {
		t.Errorf("Xxl = %d, want 48", s.Xxl)
	}
}

func TestBorderRoleDefaults(t *testing.T) {
	if defaultBorderForRole[BorderRoleContainer].Top == "" {
		t.Error("BorderRoleContainer border is empty")
	}
	if defaultBorderForRole[BorderRoleContainer].TopLeft != "┌" {
		t.Errorf("Container expected normal border, got TopLeft=%q", defaultBorderForRole[BorderRoleContainer].TopLeft)
	}
	if defaultBorderForRole[BorderRoleCard].TopLeft != "╭" {
		t.Errorf("Card expected rounded border, got TopLeft=%q", defaultBorderForRole[BorderRoleCard].TopLeft)
	}
	if defaultBorderForRole[BorderRoleModal].TopLeft != "╭" {
		t.Errorf("Modal expected rounded border, got TopLeft=%q", defaultBorderForRole[BorderRoleModal].TopLeft)
	}
	if defaultBorderForRole[BorderRoleNone].Top == "" {
		t.Errorf("BorderRoleNone should have non-empty hidden border, got Top=%q", defaultBorderForRole[BorderRoleNone].Top)
	}
}

func TestThemeBorderFor(t *testing.T) {
	th := Default
	if th.BorderFor(BorderRoleCard).TopLeft != "╭" {
		t.Errorf("BorderFor(Card) expected rounded, got TopLeft=%q", th.BorderFor(BorderRoleCard).TopLeft)
	}
	if th.BorderFor(BorderRoleContainer).TopLeft != "┌" {
		t.Errorf("BorderFor(Container) expected normal, got TopLeft=%q", th.BorderFor(BorderRoleContainer).TopLeft)
	}
	if th.BorderFor(BorderRoleModal).TopLeft != "╭" {
		t.Errorf("BorderFor(Modal) expected rounded, got TopLeft=%q", th.BorderFor(BorderRoleModal).TopLeft)
	}
	if th.BorderFor(BorderRoleNone).Top == "" {
		t.Error("BorderFor(None) should have non-empty hidden border")
	}
}

func TestThemeHasSpacing(t *testing.T) {
	th := Default
	if th.Spacing.Xxs == 0 {
		t.Error("Spacing.Xxs should not be zero")
	}
}

func TestStoreDarkLightRegistered(t *testing.T) {
	s := NewStore()
	if !s.Has("dark") {
		t.Error("expected 'dark' in new store")
	}
	if !s.Has("light") {
		t.Error("expected 'light' in new store")
	}
	if !s.Has("tokyonight") {
		t.Error("expected 'tokyonight' in new store")
	}
	th, err := s.Switch("tokyonight")
	if err != nil {
		t.Fatal(err)
	}
	if string(th.Palette.Blue) != "#7aa2f7" {
		t.Errorf("switched to tokyonight: Blue = %q, want #7aa2f7", th.Palette.Blue)
	}
}

func TestStoreDefaultStore(t *testing.T) {
	if DefaultStore.CurrentName() != "default" {
		t.Error("DefaultStore should start with 'default'")
	}
}

func TestStoreRegisterPath(t *testing.T) {
	s := NewStore()
	dir := t.TempDir()
	p := filepath.Join(dir, "dark.yaml")
	if err := os.WriteFile(p, []byte("icons: fallback"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterPath("dark", p); err != nil {
		t.Fatal(err)
	}
	if !s.Has("dark") {
		t.Error("expected 'dark' to be registered")
	}
	th, _ := s.Get("dark")
	if th.Icons.Check != "[x]" {
		t.Errorf("got %q, want [x]", th.Icons.Check)
	}
}

func TestStoreRegisterPathError(t *testing.T) {
	s := NewStore()
	err := s.RegisterPath("missing", "testdata/nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultStoreConvenience(t *testing.T) {
	// Test that top-level functions operate on DefaultStore
	prev := CurrentName()
	_ = prev

	Register("_test_convenience", NewTheme(DefaultPalette))
	t.Cleanup(func() { DefaultStore.Unregister("_test_convenience") })

	names := Names()
	found := false
	for _, n := range names {
		if n == "_test_convenience" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected '_test_convenience' in Names()")
	}
}

func TestLoadThemeWithOSEnv(t *testing.T) {
	// Create a temp file to test LoadTheme reads from actual filesystem
	dir := t.TempDir()
	p := filepath.Join(dir, "theme.yaml")
	if err := os.WriteFile(p, []byte("icons: fallback"), 0644); err != nil {
		t.Fatal(err)
	}
	th, err := LoadTheme(p)
	if err != nil {
		t.Fatal(err)
	}
	if th.Icons.Check != "[x]" {
		t.Errorf("got %q, want [x]", th.Icons.Check)
	}
}
