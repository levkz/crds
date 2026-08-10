-- +goose Up
ALTER TABLE decks ADD COLUMN input_mappings TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE decks DROP COLUMN input_mappings;
