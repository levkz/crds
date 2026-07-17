package model

type Entry struct {
	ID           string        `yaml:"id"`
	Term         string        `yaml:"term"`
	Translations []Translation `yaml:"translations"`
	Examples     []Example     `yaml:"examples"`
	Tags         []string      `yaml:"tags"`
	Notes        string        `yaml:"notes"`
}

type Translation struct {
	Text string `yaml:"text"`
}

type Example struct {
	Text        string `yaml:"text"`
	Translation string `yaml:"translation"`
}
