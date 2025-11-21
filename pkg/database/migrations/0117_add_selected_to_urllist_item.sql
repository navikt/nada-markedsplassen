-- +goose Up
ALTER TABLE workstations_url_lists ADD COLUMN selected BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE workstations_url_lists DROP COLUMN selected;