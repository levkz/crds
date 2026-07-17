package keymap

type Binding struct {
	Keys     []string
	Help     string
}

type Global struct {
	Quit Binding
	Help Binding
	Back Binding
}

type List struct {
	Up     Binding
	Down   Binding
	Select Binding
}

type Quiz struct {
	Reveal  Binding
	Again   Binding
	Hard    Binding
	Good    Binding
	Easy    Binding
}

type Search struct {
	FocusToggle Binding
	Select     Binding
	DeleteChar Binding
	List
}

var DefaultGlobal = Global{
	Quit: Binding{Keys: []string{"ctrl+c"}, Help: "quit"},
	Help: Binding{Keys: []string{"?"}, Help: "help"},
	Back: Binding{Keys: []string{"esc"}, Help: "back"},
}

var DefaultList = List{
	Up:     Binding{Keys: []string{"up", "k"}, Help: "↑"},
	Down:   Binding{Keys: []string{"down", "j"}, Help: "↓"},
	Select: Binding{Keys: []string{"enter"}, Help: "select"},
}

var DefaultQuiz = Quiz{
	Reveal: Binding{Keys: []string{"enter"}, Help: "reveal"},
	Again:  Binding{Keys: []string{"1"}, Help: "1 again"},
	Hard:   Binding{Keys: []string{"2"}, Help: "2 hard"},
	Good:   Binding{Keys: []string{"3"}, Help: "3 good"},
	Easy:   Binding{Keys: []string{"4"}, Help: "4 easy"},
}

var DefaultSearch = Search{
	FocusToggle: Binding{Keys: []string{"tab"}, Help: "focus input"},
	Select:     Binding{Keys: []string{"enter"}, Help: "open"},
	DeleteChar: Binding{Keys: []string{"backspace"}, Help: ""},
	List:       DefaultList,
}
