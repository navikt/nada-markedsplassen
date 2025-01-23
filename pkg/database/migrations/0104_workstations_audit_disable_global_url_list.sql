-- +goose Up
ALTER TABLE workstations_url_list_history ADD COLUMN disable_global_url_list BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE workstations_url_list_history DROP COLUMN disable_global_url_list;
