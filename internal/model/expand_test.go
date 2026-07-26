package model

import (
	"reflect"
	"testing"
)

func TestExpandVariants(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		expect []string
	}{
		{
			name:   "no slots",
			text:   "eat",
			expect: []string{"eat"},
		},
		{
			name:   "empty string",
			text:   "",
			expect: []string{""},
		},
		{
			name:   "single optional group",
			text:   "(to) eat",
			expect: []string{"eat", "to eat"},
		},
		{
			name:   "optional group with trailing space",
			text:   "(to )eat",
			expect: []string{"eat", "to eat"},
		},
		{
			name:   "mid-word optional",
			text:   "eat(s)",
			expect: []string{"eat", "eats"},
		},
		{
			name:   "prefix optional",
			text:   "(un)necessary",
			expect: []string{"necessary", "unnecessary"},
		},
		{
			name:   "alternatives optional",
			text:   "(he/she/we) eat",
			expect: []string{"eat", "he eat", "she eat", "we eat"},
		},
		{
			name:   "required group",
			text:   "[he] eats",
			expect: []string{"he eats"},
		},
		{
			name:   "required alternatives",
			text:   "[he/she] eats",
			expect: []string{"he eats", "she eats"},
		},
		{
			name:   "mixed optional required",
			text:   "(a)[b/c]",
			expect: []string{"b", "c", "ab", "ac"},
		},
		{
			name:   "optional with required",
			text:   "(he/she) [must] eat",
			expect: []string{"must eat", "he must eat", "she must eat"},
		},
		{
			name:   "two optional groups",
			text:   "(a)(b)",
			expect: []string{"", "b", "a", "ab"},
		},
		{
			name:   "no collapse",
			text:   "plain text",
			expect: []string{"plain text"},
		},
		{
			name:   "only slashes no parens",
			text:   "il/elle/on",
			expect: []string{"il/elle/on"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Translation{Text: tt.text}
			got := tr.ExpandVariants()
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("ExpandVariants(%q) = %v, want %v", tt.text, got, tt.expect)
			}
		})
	}
}

func TestParseSlots(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		expect []slot
	}{
		{
			name: "optional single",
			text: "(to) eat",
			expect: []slot{
				{start: 0, end: 4, content: "to", alternatives: []string{"to"}, required: false},
			},
		},
		{
			name: "alternatives optional",
			text: "(he/she) eat(s)",
			expect: []slot{
				{start: 0, end: 8, content: "he/she", alternatives: []string{"he", "she"}, required: false},
				{start: 12, end: 15, content: "s", alternatives: []string{"s"}, required: false},
			},
		},
		{
			name: "required alternatives",
			text: "[he/she]",
			expect: []slot{
				{start: 0, end: 8, content: "he/she", alternatives: []string{"he", "she"}, required: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSlots(tt.text)
			if len(got) != len(tt.expect) {
				t.Fatalf("parseSlots(%q) = %d slots, want %d: %+v", tt.text, len(got), len(tt.expect), got)
			}
			for i := range got {
				if got[i].start != tt.expect[i].start ||
					got[i].end != tt.expect[i].end ||
					got[i].content != tt.expect[i].content ||
					got[i].required != tt.expect[i].required ||
					!reflect.DeepEqual(got[i].alternatives, tt.expect[i].alternatives) {
					t.Errorf("slot %d: got %+v, want %+v", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestCartesianProduct(t *testing.T) {
	sets := [][]string{
		{"", "a"},
		{"", "b"},
	}
	got := cartesianProduct(sets)
	expect := [][]string{
		{"", ""},
		{"", "b"},
		{"a", ""},
		{"a", "b"},
	}
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("cartesianProduct(%v) = %v, want %v", sets, got, expect)
	}
}
