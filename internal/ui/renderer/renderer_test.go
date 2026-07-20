package renderer

import (
	"reflect"
	"testing"
)

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{name: "empty", s: "", want: 0},
		{name: "ascii", s: "hello", want: 5},
		{name: "cjk", s: "你好", want: 4},
		{name: "mixed", s: "a你好b", want: 6},
		{name: "ascii with ansi", s: "\x1b[31mhello\x1b[0m", want: 5},
		{name: "cjk with ansi", s: "\x1b[1m你好\x1b[0m", want: 4},
		{name: "spaces", s: "a   b", want: 5},
		{name: "emoji", s: "👍", want: 2},
		{name: "wide chars", s: "ａｂ", want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VisibleWidth(tt.s); got != tt.want {
				t.Errorf("VisibleWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestLineWidth(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{name: "empty", s: "", want: 0},
		{name: "ascii", s: "hello", want: 5},
		{name: "ansi", s: "\x1b[32mok\x1b[0m", want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LineWidth(tt.s); got != tt.want {
				t.Errorf("LineWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestMaxLineWidth(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{name: "empty", text: "", want: 0},
		{name: "single line", text: "hello", want: 5},
		{name: "multiple lines", text: "a\nbbb\ncc", want: 3},
		{name: "trailing newline", text: "aaa\n", want: 3},
		{name: "ansi lines", text: "\x1b[31mhi\x1b[0m\n\x1b[32m!\x1b[0m", want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxLineWidth(tt.text); got != tt.want {
				t.Errorf("MaxLineWidth(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

func TestTextDimensions(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantW     int
		wantH     int
	}{
		{name: "empty", text: "", wantW: 0, wantH: 0},
		{name: "single line", text: "hello", wantW: 5, wantH: 1},
		{name: "two lines", text: "hi\nbye", wantW: 3, wantH: 2},
		{name: "uneven", text: "a\nlonger\nc", wantW: 6, wantH: 3},
		{name: "ansi", text: "\x1b[31mhi\x1b[0m\n\x1b[32m!\x1b[0m", wantW: 2, wantH: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := TextDimensions(tt.text)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("TextDimensions(%q) = (%d,%d), want (%d,%d)", tt.text, gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{name: "empty", s: "", want: ""},
		{name: "no ansi", s: "hello", want: "hello"},
		{name: "single sgr", s: "\x1b[31mhello\x1b[0m", want: "hello"},
		{name: "bold", s: "\x1b[1mbold\x1b[0m", want: "bold"},
		{name: "multiple params", s: "\x1b[1;32mok\x1b[0m", want: "ok"},
		{name: "non-sgr csi", s: "\x1b[?25lhide\x1b[?25h", want: "hide"},
		{name: "adjacent", s: "\x1b[31m\x1b[1mtext\x1b[0m", want: "text"},
		{name: "no trailing reset", s: "\x1b[31mhello", want: "hello"},
		{name: "bare escape", s: "a\x1bb", want: "ab"},
		{name: "escape at end", s: "hello\x1b", want: "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripANSI(tt.s); got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestCountANSISequences(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{name: "empty", s: "", want: 0},
		{name: "no ansi", s: "hello", want: 0},
		{name: "one sgr", s: "\x1b[31mhello", want: 1},
		{name: "open and close", s: "\x1b[31mhello\x1b[0m", want: 2},
		{name: "bold and color", s: "\x1b[1m\x1b[32mok\x1b[0m", want: 3},
		{name: "non-sgr", s: "\x1b[?25l", want: 1},
		{name: "bare escape", s: "a\x1bb", want: 0},
		{name: "c1 control", s: "\x1bM", want: 1},
		{name: "mixed valid and stray", s: "\x1b[31ma\x1bb", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountANSISequences(tt.s); got != tt.want {
				t.Errorf("CountANSISequences(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     []string
	}{
		{name: "empty", text: "", maxWidth: 10, want: nil},
		{name: "fits", text: "hello", maxWidth: 10, want: []string{"hello"}},
		{name: "exact", text: "hello", maxWidth: 5, want: []string{"hello"}},
		{name: "two words wrap", text: "hello world", maxWidth: 8, want: []string{"hello", "world"}},
		{name: "three words wrap", text: "a bb ccc", maxWidth: 4, want: []string{"a bb", "ccc"}},
		{name: "long word", text: "superlongword", maxWidth: 5, want: []string{"super", "longw", "ord"}},
		{name: "multi line", text: "hello world\nfoo bar baz", maxWidth: 8, want: []string{"hello", "world", "foo bar", "baz"}},
		{name: "single char width", text: "ab cd", maxWidth: 1, want: []string{"a", "b", "c", "d"}},
		{name: "zero width", text: "hello", maxWidth: 0, want: []string{"h", "e", "l", "l", "o"}},
		{name: "negative width", text: "hi", maxWidth: -1, want: []string{"h", "i"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.text, tt.maxWidth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Wrap(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{name: "empty", text: "", maxWidth: 10, want: ""},
		{name: "fits", text: "hello", maxWidth: 10, want: "hello"},
		{name: "exact", text: "hello", maxWidth: 5, want: "hello"},
		{name: "truncate", text: "hello world", maxWidth: 8, want: "hello w…"},
		{name: "truncate short", text: "hello world", maxWidth: 3, want: "he…"},
		{name: "zero width", text: "hello", maxWidth: 0, want: ""},
		{name: "two chars", text: "hello", maxWidth: 2, want: "h…"},
		{name: "ansi", text: "\x1b[31mhello world\x1b[0m", maxWidth: 8, want: "hello w…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.text, tt.maxWidth); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestFit(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{name: "empty", text: "", maxWidth: 10, want: ""},
		{name: "fits", text: "hello", maxWidth: 10, want: "hello"},
		{name: "exact", text: "hello", maxWidth: 5, want: "hello"},
		{name: "truncate", text: "hello world", maxWidth: 8, want: "hello wo"},
		{name: "short", text: "hello world", maxWidth: 3, want: "hel"},
		{name: "zero width", text: "hello", maxWidth: 0, want: ""},
		{name: "ansi", text: "\x1b[31mhello world\x1b[0m", maxWidth: 8, want: "hello wo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Fit(tt.text, tt.maxWidth); got != tt.want {
				t.Errorf("Fit(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}
