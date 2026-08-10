package fuzzy

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"already clean", "hello", "hello"},
		{"mixed case", "Hello World", "hello world"},
		{"multiple spaces", "hello   world", "hello world"},
		{"leading trailing", "  hello world  ", "hello world"},
		{"punctuation", "hello, world!", "hello, world!"},
		{"tabs newlines", "hello\t\nworld", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.input); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLevenshteinSimilarity(t *testing.T) {
	m := &LevenshteinMatcher{}
	tests := []struct {
		name string
		a, b string
		want float64
	}{
		{"identical", "hello", "hello", 1.0},
		{"both empty", "", "", 1.0},
		{"one empty a", "", "hello", 0.0},
		{"one empty b", "hello", "", 0.0},
		{"completely different", "abc", "xyz", 0.0},
		{"close match", "kitten", "sitten", 1.0 - 1.0/6.0},
		{"unicode", "café", "cafe", 0.75},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.Similarity(tt.a, tt.b)
			if !approxEqual(got, tt.want, 1e-9) {
				t.Errorf("Similarity(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func approxEqual(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= eps
}

func TestFuzzyMatcherCheck(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		score, matched := fm.Check("hello", []string{"hello"})
		if !matched || score != 1.0 {
			t.Errorf("expected matched=true, score=1.0, got matched=%v, score=%v", matched, score)
		}
	})

	t.Run("close match", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		score, matched := fm.Check("kitten", []string{"sitten"})
		if !matched {
			t.Errorf("expected matched=true, got matched=%v, score=%v", matched, score)
		}
	})

	t.Run("no match", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		score, matched := fm.Check("abc", []string{"xyz"})
		if matched {
			t.Errorf("expected matched=false, got matched=%v, score=%v", matched, score)
		}
		_ = score
	})

	t.Run("empty answers", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		score, matched := fm.Check("hello", nil)
		if matched {
			t.Errorf("expected matched=false, got matched=%v", matched)
		}
		if score != 0.0 {
			t.Errorf("expected score=0.0, got %v", score)
		}
	})

	t.Run("normalization applied", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		score, matched := fm.Check("  Hello  ", []string{"hello"})
		if !matched || score != 1.0 {
			t.Errorf("expected matched=true, score=1.0, got matched=%v, score=%v", matched, score)
		}
	})
}

func TestFuzzyMatcherGrade(t *testing.T) {
	t.Run("exact match returns Good", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		if got := fm.Grade("hello", []string{"hello"}); got != Good {
			t.Errorf("expected %d, got %d", Good, got)
		}
	})

	t.Run("exact match after normalization", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		if got := fm.Grade("  Hello  ", []string{"hello"}); got != Good {
			t.Errorf("expected %d, got %d", Good, got)
		}
	})

	t.Run("close match returns Hard", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		if got := fm.Grade("kitten", []string{"sitten"}); got != Hard {
			t.Errorf("expected %d, got %d", Hard, got)
		}
	})

	t.Run("no match returns Again", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		if got := fm.Grade("abc", []string{"xyz"}); got != Again {
			t.Errorf("expected %d, got %d", Again, got)
		}
	})

	t.Run("empty answers returns Again", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.7)
		if got := fm.Grade("hello", nil); got != Again {
			t.Errorf("expected %d, got %d", Again, got)
		}
	})

	t.Run("exact match preferred over high similarity", func(t *testing.T) {
		fm := NewFuzzyMatcher(0.3)
		if got := fm.Grade("hello", []string{"hallo", "hello"}); got != Good {
			t.Errorf("expected %d, got %d", Good, got)
		}
	})
}

func TestNewFuzzyMatcherDefaultThreshold(t *testing.T) {
	fm := NewFuzzyMatcher(0)
	if fm.Threshold != 0.7 {
		t.Errorf("expected default threshold 0.7, got %v", fm.Threshold)
	}
}

func TestNewFuzzyMatcherNegativeThreshold(t *testing.T) {
	fm := NewFuzzyMatcher(-1)
	if fm.Threshold != 0.7 {
		t.Errorf("expected default threshold 0.7, got %v", fm.Threshold)
	}
}

func TestStripAccents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no accents", "hello", "hello"},
		{"french accents", "café crème déjà", "cafe creme deja"},
		{"multiple on one letter", "cafèêé", "cafeee"},
		{"umlauts", "für über grüße", "fur uber gruße"},
		{"spanish", "mañana corazón", "manana corazon"},
		{"portuguese", "coração você", "coracao voce"},
		{"non-decomposing kept", "straße øre", "straße øre"},
		{"mixed", "Élève naïf", "Eleve naif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripAccents(tt.input); got != tt.want {
				t.Errorf("StripAccents(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Mode
	}{
		{"empty defaults approximate", "", Approximate},
		{"whitespace defaults approximate", "  ", Approximate},
		{"strict", "strict", Strict},
		{"strict uppercase", "STRICT", Strict},
		{"strict padded", "  strict ", Strict},
		{"approximate", "approximate", Approximate},
		{"unknown defaults approximate", "bogus", Approximate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseMode(tt.in); got != tt.want {
				t.Errorf("ParseMode(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestModeString(t *testing.T) {
	if Strict.String() != "strict" {
		t.Errorf("Strict.String() = %q, want %q", Strict.String(), "strict")
	}
	if Approximate.String() != "approximate" {
		t.Errorf("Approximate.String() = %q, want %q", Approximate.String(), "approximate")
	}
}

func TestFuzzyMatcherApproximate(t *testing.T) {
	fm := NewFuzzyMatcher(0.7)
	fm.Mode = Approximate

	t.Run("accented answer, plain input is exact", func(t *testing.T) {
		if got := fm.Grade("cafe", []string{"café"}); got != Good {
			t.Errorf("expected Good, got %d", got)
		}
	})

	t.Run("accented input, plain answer is exact", func(t *testing.T) {
		if got := fm.Grade("café", []string{"cafe"}); got != Good {
			t.Errorf("expected Good, got %d", got)
		}
	})

	t.Run("accented input, accented answer is exact", func(t *testing.T) {
		if got := fm.Grade("café", []string{"café"}); got != Good {
			t.Errorf("expected Good, got %d", got)
		}
	})

	t.Run("accent still differentiates words", func(t *testing.T) {
		if got := fm.Grade("maison", []string{"café"}); got != Again {
			t.Errorf("expected Again, got %d", got)
		}
	})

	t.Run("check matches plain input to accented answer", func(t *testing.T) {
		score, matched := fm.Check("cafe", []string{"café"})
		if !matched || score != 1.0 {
			t.Errorf("expected matched=true score=1.0, got matched=%v score=%v", matched, score)
		}
	})
}

func TestFuzzyMatcherStrict(t *testing.T) {
	fm := NewFuzzyMatcher(0.7)
	fm.Mode = Strict

	t.Run("plain input vs accented answer is not exact", func(t *testing.T) {
		got := fm.Grade("cafe", []string{"café"})
		if got == Good {
			t.Errorf("expected not Good, got %d", got)
		}
	})

	t.Run("accented input vs accented answer is exact", func(t *testing.T) {
		if got := fm.Grade("café", []string{"café"}); got != Good {
			t.Errorf("expected Good, got %d", got)
		}
	})
}

func TestNewFuzzyMatcherDefaultMode(t *testing.T) {
	fm := NewFuzzyMatcher(0)
	if fm.Mode != Approximate {
		t.Errorf("expected default mode Approximate, got %v", fm.Mode)
	}
}
