package parser

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"

	"crds/internal/model"
)

// ParseFile reads, parses, validates and normalizes a vocabulary deck.
func ParseFile(path string) (*model.Deck, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	deck, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", path, err)
	}

	return deck, nil
}

// Parse parses a deck from YAML bytes.
func Parse(data []byte) (*model.Deck, error) {
	var deck model.Deck

	if err := yaml.Unmarshal(data, &deck); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	Normalize(&deck)
	assignIDs(&deck)

	if err := Validate(&deck); err != nil {
		return nil, err
	}

	return &deck, nil
}
