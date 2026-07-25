package styles

import (
	"strings"
	"testing"

	"crds/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

func testStyleRender(t *testing.T, name string, s lipgloss.Style) {
	t.Helper()
	result := s.Render("test")
	if result == "" {
		t.Errorf("%s.Render returned empty", name)
	}
	if !strings.Contains(result, "test") {
		t.Errorf("%s.Render does not contain input text", name)
	}
}

func TestHeader(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Header(60)
		testStyleRender(t, "Header", s)
	})
	t.Run("width", func(t *testing.T) {
		s := Header(80)
		if w := s.GetWidth(); w != 80 {
			t.Errorf("Header width = %d, want 80", w)
		}
	})
	t.Run("padding", func(t *testing.T) {
		s := Header(60)
		top, right, bottom, left := s.GetPadding()
		if top != 0 || bottom != 0 {
			t.Errorf("Header vertical padding = (%d, %d), want (0, 0)", top, bottom)
		}
		if right != 1 || left != 1 {
			t.Errorf("Header horizontal padding = (%d, %d), want (1, 1)", right, left)
		}
	})
	t.Run("bold", func(t *testing.T) {
		s := Header(60)
		if !s.GetBold() {
			t.Error("Header should be bold")
		}
	})
	t.Run("background", func(t *testing.T) {
		s := Header(60)
		bg := s.GetBackground()
		if bg != ui.Theme.Palette.Surface {
			t.Errorf("Header background = %v, want %v", bg, ui.Theme.Palette.Surface)
		}
	})
}

func TestFooter(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Footer(60)
		testStyleRender(t, "Footer", s)
	})
	t.Run("width", func(t *testing.T) {
		s := Footer(80)
		if w := s.GetWidth(); w != 80 {
			t.Errorf("Footer width = %d, want 80", w)
		}
	})
	t.Run("padding", func(t *testing.T) {
		s := Footer(60)
		_, right, _, left := s.GetPadding()
		if right != 1 || left != 1 {
			t.Errorf("Footer horizontal padding = (%d, %d), want (1, 1)", right, left)
		}
	})
	t.Run("color", func(t *testing.T) {
		s := Footer(60)
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Gray {
			t.Errorf("Footer foreground = %v, want %v", fg, ui.Theme.Palette.Gray)
		}
	})
	t.Run("background", func(t *testing.T) {
		s := Footer(60)
		bg := s.GetBackground()
		if bg != ui.Theme.Palette.Surface {
			t.Errorf("Footer background = %v, want %v", bg, ui.Theme.Palette.Surface)
		}
	})
}

func TestSelectedItem(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := SelectedItem()
		testStyleRender(t, "SelectedItem", s)
	})
	t.Run("background", func(t *testing.T) {
		s := SelectedItem()
		bg := s.GetBackground()
		if bg != ui.Theme.Palette.Selection {
			t.Errorf("SelectedItem background = %v, want %v", bg, ui.Theme.Palette.Selection)
		}
	})
	t.Run("foreground", func(t *testing.T) {
		s := SelectedItem()
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Blue {
			t.Errorf("SelectedItem foreground = %v, want %v", fg, ui.Theme.Palette.Blue)
		}
	})
	t.Run("chainable", func(t *testing.T) {
		s := SelectedItem().Width(40)
		if w := s.GetWidth(); w != 40 {
			t.Errorf("SelectedItem chained width = %d, want 40", w)
		}
	})
}

func TestFocusedInput(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := FocusedInput()
		testStyleRender(t, "FocusedInput", s)
	})
	t.Run("border", func(t *testing.T) {
		s := FocusedInput()
		bdr := s.GetBorderStyle()
		if bdr.Top == "" || bdr.Bottom == "" || bdr.Left == "" || bdr.Right == "" {
			t.Error("FocusedInput should have non-empty border characters")
		}
	})
	t.Run("border rounded", func(t *testing.T) {
		s := FocusedInput()
		bdr := s.GetBorderStyle()
		if bdr.TopLeft != "╭" {
			t.Errorf("FocusedInput expected rounded border, got TopLeft=%q", bdr.TopLeft)
		}
	})
	t.Run("border foreground", func(t *testing.T) {
		s := FocusedInput()
		bf := s.GetBorderTopForeground()
		if bf != ui.Theme.Palette.Blue {
			t.Errorf("FocusedInput border foreground = %v, want %v", bf, ui.Theme.Palette.Blue)
		}
	})
	t.Run("padding", func(t *testing.T) {
		s := FocusedInput()
		_, right, _, left := s.GetPadding()
		if right != 1 || left != 1 {
			t.Errorf("FocusedInput horizontal padding = (%d, %d), want (1, 1)", right, left)
		}
	})
}

func TestError(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Error()
		testStyleRender(t, "Error", s)
	})
	t.Run("color", func(t *testing.T) {
		s := Error()
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Red {
			t.Errorf("Error foreground = %v, want %v", fg, ui.Theme.Palette.Red)
		}
	})
	t.Run("chainable", func(t *testing.T) {
		s := Error().Bold(true)
		if !s.GetBold() {
			t.Error("Error chained bold not applied")
		}
	})
}

func TestWarning(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Warning()
		testStyleRender(t, "Warning", s)
	})
	t.Run("color", func(t *testing.T) {
		s := Warning()
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Orange {
			t.Errorf("Warning foreground = %v, want %v", fg, ui.Theme.Palette.Orange)
		}
	})
	t.Run("chainable", func(t *testing.T) {
		s := Warning().Italic(true)
		if !s.GetItalic() {
			t.Error("Warning chained italic not applied")
		}
	})
}

func TestSuccess(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Success()
		testStyleRender(t, "Success", s)
	})
	t.Run("color", func(t *testing.T) {
		s := Success()
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Green {
			t.Errorf("Success foreground = %v, want %v", fg, ui.Theme.Palette.Green)
		}
	})
	t.Run("chainable", func(t *testing.T) {
		s := Success().Bold(true)
		if !s.GetBold() {
			t.Error("Success chained bold not applied")
		}
	})
}

func TestHint(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Hint()
		testStyleRender(t, "Hint", s)
	})
	t.Run("italic", func(t *testing.T) {
		s := Hint()
		if !s.GetItalic() {
			t.Error("Hint should be italic")
		}
	})
	t.Run("color", func(t *testing.T) {
		s := Hint()
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Gray {
			t.Errorf("Hint foreground = %v, want %v", fg, ui.Theme.Palette.Gray)
		}
	})
	t.Run("chainable", func(t *testing.T) {
		s := Hint().Width(40)
		if w := s.GetWidth(); w != 40 {
			t.Errorf("Hint chained width = %d, want 40", w)
		}
	})
}

func TestMutedText(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := MutedText()
		testStyleRender(t, "MutedText", s)
	})
	t.Run("color", func(t *testing.T) {
		s := MutedText()
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Gray {
			t.Errorf("MutedText foreground = %v, want %v", fg, ui.Theme.Palette.Gray)
		}
	})
	t.Run("no bold", func(t *testing.T) {
		s := MutedText()
		if s.GetBold() {
			t.Error("MutedText should not be bold by default")
		}
	})
	t.Run("chainable", func(t *testing.T) {
		s := MutedText().Bold(true)
		if !s.GetBold() {
			t.Error("MutedText chained bold not applied")
		}
	})
}

func TestCard(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Card(60)
		testStyleRender(t, "Card", s)
	})
	t.Run("width", func(t *testing.T) {
		s := Card(40)
		if w := s.GetWidth(); w != 40 {
			t.Errorf("Card width = %d, want 40", w)
		}
	})
	t.Run("padding", func(t *testing.T) {
		s := Card(60)
		top, right, bottom, left := s.GetPadding()
		if top != 1 || bottom != 1 {
			t.Errorf("Card vertical padding = (%d, %d), want (1, 1)", top, bottom)
		}
		if right != 1 || left != 1 {
			t.Errorf("Card horizontal padding = (%d, %d), want (1, 1)", right, left)
		}
	})
	t.Run("color", func(t *testing.T) {
		s := Card(60)
		fg := s.GetForeground()
		if fg != ui.Theme.Palette.Blue {
			t.Errorf("Card foreground = %v, want %v", fg, ui.Theme.Palette.Blue)
		}
	})
}

func TestPanel(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Panel(60)
		testStyleRender(t, "Panel", s)
	})
	t.Run("width", func(t *testing.T) {
		s := Panel(80)
		if w := s.GetWidth(); w != 80 {
			t.Errorf("Panel width = %d, want 80", w)
		}
	})
	t.Run("border", func(t *testing.T) {
		s := Panel(60)
		bdr := s.GetBorderStyle()
		if bdr.Top == "" || bdr.Bottom == "" || bdr.Left == "" || bdr.Right == "" {
			t.Error("Panel should have non-empty border characters")
		}
	})
	t.Run("border normal", func(t *testing.T) {
		s := Panel(60)
		bdr := s.GetBorderStyle()
		if bdr.TopLeft != "┌" {
			t.Errorf("Panel expected normal border, got TopLeft=%q", bdr.TopLeft)
		}
	})
	t.Run("border foreground", func(t *testing.T) {
		s := Panel(60)
		bf := s.GetBorderTopForeground()
		if bf != ui.Theme.Palette.Border {
			t.Errorf("Panel border foreground = %v, want %v", bf, ui.Theme.Palette.Border)
		}
	})
	t.Run("padding", func(t *testing.T) {
		s := Panel(60)
		_, right, _, left := s.GetPadding()
		if right != 2 || left != 2 {
			t.Errorf("Panel horizontal padding = (%d, %d), want (2, 2)", right, left)
		}
	})
}

func TestModal(t *testing.T) {
	t.Run("render", func(t *testing.T) {
		s := Modal(40, 10)
		testStyleRender(t, "Modal", s)
	})
	t.Run("width", func(t *testing.T) {
		s := Modal(50, 10)
		if w := s.GetWidth(); w != 50 {
			t.Errorf("Modal width = %d, want 50", w)
		}
	})
	t.Run("height", func(t *testing.T) {
		s := Modal(40, 15)
		if h := s.GetHeight(); h != 15 {
			t.Errorf("Modal height = %d, want 15", h)
		}
	})
	t.Run("border", func(t *testing.T) {
		s := Modal(40, 10)
		bdr := s.GetBorderStyle()
		if bdr.Top == "" || bdr.Bottom == "" || bdr.Left == "" || bdr.Right == "" {
			t.Error("Modal should have non-empty border characters")
		}
	})
	t.Run("border rounded", func(t *testing.T) {
		s := Modal(40, 10)
		bdr := s.GetBorderStyle()
		if bdr.TopLeft != "╭" {
			t.Errorf("Modal expected rounded border, got TopLeft=%q", bdr.TopLeft)
		}
	})
	t.Run("border foreground", func(t *testing.T) {
		s := Modal(40, 10)
		bf := s.GetBorderTopForeground()
		if bf != ui.Theme.Palette.Blue {
			t.Errorf("Modal border foreground = %v, want %v", bf, ui.Theme.Palette.Blue)
		}
	})
	t.Run("padding", func(t *testing.T) {
		s := Modal(40, 10)
		_, right, _, left := s.GetPadding()
		if right != 2 || left != 2 {
			t.Errorf("Modal horizontal padding = (%d, %d), want (2, 2)", right, left)
		}
	})
}

func TestAllStylesRender(t *testing.T) {
	styleFuncs := []struct {
		name string
		s    lipgloss.Style
	}{
		{"Header(60)", Header(60)},
		{"Footer(60)", Footer(60)},
		{"SelectedItem", SelectedItem()},
		{"FocusedInput", FocusedInput()},
		{"Error", Error()},
		{"Warning", Warning()},
		{"Success", Success()},
		{"Hint", Hint()},
		{"MutedText", MutedText()},
		{"Card(60)", Card(60)},
		{"Panel(60)", Panel(60)},
		{"Modal(40,10)", Modal(40, 10)},
	}
	for _, st := range styleFuncs {
		t.Run(st.name, func(t *testing.T) {
			testStyleRender(t, st.name, st.s)
		})
	}
}

func TestThemeSwitchUpdatesStyles(t *testing.T) {
	orig := ui.Theme
	t.Cleanup(func() { ui.SetTheme(orig) })

	th := ui.Theme
	th.Success = th.Success.Bold(true)
	ui.SetTheme(th)

	s := Success()
	if !s.GetBold() {
		t.Error("Success style should reflect theme switch (bold)")
	}
}
