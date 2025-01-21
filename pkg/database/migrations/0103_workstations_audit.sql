-- +goose Up
CREATE TABLE workstations_config_history (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    nav_ident TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    workstation_config JSONB NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE workstations_onprem_allowlist_history (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    nav_ident TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hosts TEXT[] NOT NULL DEFAULT '{}',
    PRIMARY KEY (id)
);

CREATE TABLE workstations_url_list_history (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    nav_ident TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    url_list TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

-- +goose Down
DROP TABLE workstations_url_list_history;
DROP TABLE workstations_onprem_allowlist_history;
DROP TABLE workstations_config_history;
