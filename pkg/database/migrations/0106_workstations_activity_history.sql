-- +goose Up
CREATE TABLE workstations_activity_history (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    nav_ident TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action TEXT NOT NULL CHECK (action IN ('START', 'STOP')),
    instance_id TEXT NOT NULL,
    PRIMARY KEY (id)
);

-- +goose Down
DROP TABLE workstations_activity_history;
