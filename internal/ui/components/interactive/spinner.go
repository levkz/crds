package components

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type spinnerTickMsg time.Time

type SpinnerModel struct {
	frame    int
	active   bool
	interval time.Duration
	frames   []string
}

var defaultSpinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func NewSpinner() SpinnerModel {
	frames := make([]string, len(defaultSpinnerFrames))
	copy(frames, defaultSpinnerFrames)
	return SpinnerModel{
		frame:    0,
		active:   false,
		interval: 80 * time.Millisecond,
		frames:   frames,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.tick()
}

func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	switch msg.(type) {
	case spinnerTickMsg:
		if m.active {
			m.frame = (m.frame + 1) % len(m.frames)
		}
		return m, m.tick()
	}
	return m, nil
}

func (m SpinnerModel) View(width int) string {
	if !m.active {
		return ""
	}
	return m.frames[m.frame%len(m.frames)]
}

func (m SpinnerModel) ViewWithMessage(msg string, width int) string {
	if !m.active {
		return msg
	}
	var b strings.Builder
	b.WriteString(m.frames[m.frame%len(m.frames)])
	b.WriteString(" ")
	b.WriteString(msg)
	return b.String()
}

func (m SpinnerModel) tick() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg {
		return spinnerTickMsg(t)
	})
}

func (m *SpinnerModel) Start() {
	m.active = true
	m.frame = 0
}

func (m *SpinnerModel) Stop()      { m.active = false }
func (m *SpinnerModel) Active() bool { return m.active }
func (m *SpinnerModel) Toggle()    { m.active = !m.active }

func (m *SpinnerModel) SetFrames(frames []string) {
	if len(frames) > 0 {
		m.frames = make([]string, len(frames))
		copy(m.frames, frames)
	}
}

func (m *SpinnerModel) SetInterval(d time.Duration) {
	if d > 0 {
		m.interval = d
	}
}
