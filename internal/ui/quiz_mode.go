package ui

type QuizMode int

const (
	QuizModeNormal QuizMode = iota
	QuizModeRandom
	QuizModeSmart
	QuizModeKindaSmart
	QuizModeDue
	NumQuizModes
)

func (m QuizMode) String() string {
	switch m {
	case QuizModeNormal:
		return "normal"
	case QuizModeRandom:
		return "random"
	case QuizModeSmart:
		return "smart"
	case QuizModeKindaSmart:
		return "kinda-smart"
	case QuizModeDue:
		return "due"
	default:
		return "normal"
	}
}

func (m QuizMode) Next() QuizMode {
	return (m + 1) % NumQuizModes
}

func ParseQuizMode(s string) QuizMode {
	switch s {
	case "random":
		return QuizModeRandom
	case "smart":
		return QuizModeSmart
	case "kinda-smart":
		return QuizModeKindaSmart
	case "due":
		return QuizModeDue
	default:
		return QuizModeNormal
	}
}
