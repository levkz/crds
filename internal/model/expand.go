package model

import (
	"strings"
)

type slot struct {
	content      string
	alternatives []string
	start        int
	end          int
	required     bool
}

func findClose(text string, open int, closeChar byte) int {
	for i := open + 1; i < len(text); i++ {
		if text[i] == closeChar {
			return i
		}
	}
	return -1
}

func parseSlots(text string) []slot {
	var slots []slot
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(':
			close := findClose(text, i, ')')
			if close == -1 {
				continue
			}
			content := text[i+1 : close]
			slots = append(slots, slot{
				start:        i,
				end:          close + 1,
				content:      content,
				alternatives: strings.Split(content, "/"),
				required:     false,
			})
			i = close
		case '[':
			close := findClose(text, i, ']')
			if close == -1 {
				continue
			}
			content := text[i+1 : close]
			slots = append(slots, slot{
				start:        i,
				end:          close + 1,
				content:      content,
				alternatives: strings.Split(content, "/"),
				required:     true,
			})
			i = close
		}
	}
	return slots
}

func contributions(slots []slot) [][]string {
	sets := make([][]string, len(slots))
	for i, s := range slots {
		if s.required {
			sets[i] = s.alternatives
		} else {
			sets[i] = append([]string{""}, s.alternatives...)
		}
	}
	return sets
}

func cartesianProduct(sets [][]string) [][]string {
	if len(sets) == 0 {
		return [][]string{{}}
	}
	result := [][]string{{}}
	for _, set := range sets {
		var next [][]string
		for _, existing := range result {
			for _, item := range set {
				combo := make([]string, len(existing)+1)
				copy(combo, existing)
				combo[len(existing)] = item
				next = append(next, combo)
			}
		}
		result = next
	}
	return result
}

func ExpandText(text string) []string {
	return Translation{Text: text}.ExpandVariants()
}

func (t Translation) ExpandVariants() []string {
	if t.Text == "" {
		return []string{""}
	}

	slots := parseSlots(t.Text)
	if len(slots) == 0 {
		return []string{t.Text}
	}

	sets := contributions(slots)
	combos := cartesianProduct(sets)
	seen := make(map[string]struct{}, len(combos))
	result := make([]string, 0, len(combos))

	for _, combo := range combos {
		var sb strings.Builder
		lastEnd := 0
		for i, s := range slots {
			if s.start > lastEnd {
				sb.WriteString(t.Text[lastEnd:s.start])
			}
			sb.WriteString(combo[i])
			lastEnd = s.end
		}
		if lastEnd < len(t.Text) {
			sb.WriteString(t.Text[lastEnd:])
		}
		variant := strings.Join(strings.Fields(sb.String()), " ")
		if _, ok := seen[variant]; !ok {
			seen[variant] = struct{}{}
			result = append(result, variant)
		}
	}
	return result
}
