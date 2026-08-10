package model

type Deck struct {
	ID                  string            `yaml:"id"`
	Name                string            `yaml:"name"`
	Language            string            `yaml:"language"`
	TranslationLanguage string            `yaml:"translation_language"`
	InputMappings       map[string]string `yaml:"input_mappings,omitempty"`
	Entries             []Entry           `yaml:"entries"`
}
