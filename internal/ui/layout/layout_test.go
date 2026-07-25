package layout

import (
	"testing"

	"crds/internal/ui/renderer"
)

func TestPage(t *testing.T) {
	tests := []struct {
		name   string
		header string
		body   string
		footer string
		height int
		want   string
	}{
		{
			name:   "basic",
			header: "header",
			body:   "body",
			footer: "footer",
			height: 0,
			want:   "header\n\nbody\n\nfooter",
		},
		{
			name:   "empty header",
			header: "",
			body:   "body",
			footer: "footer",
			height: 0,
			want:   "\n\nbody\n\nfooter",
		},
		{
			name:   "empty body",
			header: "header",
			body:   "",
			footer: "footer",
			height: 0,
			want:   "header\n\n\n\nfooter",
		},
		{
			name:   "empty footer",
			header: "header",
			body:   "body",
			footer: "",
			height: 0,
			want:   "header\n\nbody\n\n",
		},
		{
			name:   "anchors footer to bottom",
			header: "h",
			body:   "b",
			footer: "f",
			height: 10,
			want:   "h\n\nb\n\n\n\n\n\n\nf",
		},
		{
			name:   "no padding when body fits exactly",
			header: "h",
			body:   "b",
			footer: "f",
			height: 5,
			want:   "h\n\nb\n\nf",
		},
		{
			name:   "no padding when body exceeds height",
			header: "h",
			body:   "b\nc\nd\ne",
			footer: "f",
			height: 5,
			want:   "h\n\nb\nc\nd\ne\n\nf",
		},
		{
			name:   "anchors with multi-line footer",
			header: "h",
			body:   "b",
			footer: "f\ng",
			height: 10,
			want:   "h\n\nb\n\n\n\n\n\nf\ng",
		},
		{
			name:   "zero height falls back to no anchoring",
			header: "h",
			body:   "b",
			footer: "f",
			height: 0,
			want:   "h\n\nb\n\nf",
		},
		{
			name:   "negative height falls back to no anchoring",
			header: "h",
			body:   "b",
			footer: "f",
			height: -1,
			want:   "h\n\nb\n\nf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Page(tt.header, tt.body, tt.footer, tt.height)
			if got != tt.want {
				t.Errorf("Page() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColumn(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{
			name:  "single item",
			items: []string{"a"},
			want:  "a",
		},
		{
			name:  "two items",
			items: []string{"a", "b"},
			want:  "a\n\nb",
		},
		{
			name:  "three items",
			items: []string{"a", "b", "c"},
			want:  "a\n\nb\n\nc",
		},
		{
			name:  "empty",
			items: nil,
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Column(tt.items...)
			if got != tt.want {
				t.Errorf("Column() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVSpace(t *testing.T) {
	tests := []struct {
		name  string
		n     int
		want  string
	}{
		{name: "zero", n: 0, want: ""},
		{name: "one", n: 1, want: "\n"},
		{name: "two", n: 2, want: "\n\n"},
		{name: "three", n: 3, want: "\n\n\n"},
		{name: "negative", n: -1, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VSpace(tt.n)
			if got != tt.want {
				t.Errorf("VSpace(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestHSpace(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{name: "zero", n: 0, want: ""},
		{name: "one", n: 1, want: " "},
		{name: "four", n: 4, want: "    "},
		{name: "negative", n: -1, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HSpace(tt.n)
			if got != tt.want {
				t.Errorf("HSpace(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestCenter(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
	}{
		{name: "short", text: "x", width: 10},
		{name: "exact", text: "hello", width: 5},
		{name: "wide", text: "hello world", width: 30},
		{name: "empty", text: "", width: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Center(tt.text, tt.width)
			if renderer.VisibleWidth(got) != tt.width {
				t.Errorf("Center() returned visible width %d, want %d (got=%q)", renderer.VisibleWidth(got), tt.width, got)
			}
		})
	}
}

func TestAlignLeft(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
	}{
		{name: "short", text: "x", width: 10},
		{name: "exact", text: "hello", width: 5},
		{name: "empty", text: "", width: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlignLeft(tt.text, tt.width)
			if renderer.VisibleWidth(got) != tt.width {
				t.Errorf("AlignLeft() returned visible width %d, want %d (got=%q)", renderer.VisibleWidth(got), tt.width, got)
			}
		})
	}
}

func TestAlignRight(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
	}{
		{name: "short", text: "x", width: 10},
		{name: "exact", text: "hello", width: 5},
		{name: "empty", text: "", width: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlignRight(tt.text, tt.width)
			if renderer.VisibleWidth(got) != tt.width {
				t.Errorf("AlignRight() returned visible width %d, want %d (got=%q)", renderer.VisibleWidth(got), tt.width, got)
			}
		})
	}
}

func TestRow(t *testing.T) {
	tests := []struct {
		name  string
		items []string
	}{
		{name: "single", items: []string{"a"}},
		{name: "two", items: []string{"a", "b"}},
		{name: "three", items: []string{"a", "b", "c"}},
		{name: "empty", items: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Row(tt.items...)
			if len(tt.items) == 0 && got != "" {
				t.Errorf("Row() = %q, want empty", got)
			}
			if len(tt.items) > 0 && got == "" {
				t.Errorf("Row() = empty, want non-empty")
			}
		})
	}
}

func TestGrid(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		cols  int
		want  string
	}{
		{name: "empty", items: nil, cols: 3, want: ""},
		{name: "single col two items", items: []string{"a", "b"}, cols: 1, want: "a\nb"},
		{name: "two cols four items", items: []string{"a", "b", "c", "d"}, cols: 2, want: "ab\ncd"},
		{name: "two cols three items", items: []string{"a", "b", "c"}, cols: 2, want: "ab\nc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Grid(tt.items, tt.cols)
			if got != tt.want {
				t.Errorf("Grid() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStack(t *testing.T) {
	tests := []struct {
		name   string
		layers []string
	}{
		{name: "single layer", layers: []string{"hello"}},
		{name: "two layers", layers: []string{"X", "---"}},
		{name: "empty", layers: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Stack(tt.layers...)
			if len(tt.layers) == 0 && got != "" {
				t.Errorf("Stack() = %q, want empty", got)
			}
			if len(tt.layers) == 1 && got != tt.layers[0] {
				t.Errorf("Stack() = %q, want %q", got, tt.layers[0])
			}
		})
	}
}
