-- +goose Up
CREATE TABLE metabase_dashboard (
    "id"            uuid                 DEFAULT uuid_generate_v4(),
    "name"          TEXT        NOT NULL,
    "description"   TEXT,
    "group"         TEXT        NOT NULL,
    "public_dashboard_id"       uuid        NOT NULL,
    "metabase_id"   INT         NOT NULL UNIQUE, 
    "created_by"    TEXT        NOT NULL,
    "created"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "last_modified" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "tsv_document"  tsvector GENERATED ALWAYS AS (
                                to_tsvector('norwegian', "name")
                                || to_tsvector('norwegian', coalesce("description", ''))
                        ) STORED,
    "keywords"       TEXT[] NOT NULL DEFAULT '{}',
    "teamkatalogen_url" TEXT,
    "team_id"       uuid,
    PRIMARY KEY (id)
);

CREATE TRIGGER metabase_dashboard_set_modified
    BEFORE UPDATE
    ON metabase_dashboard
    FOR EACH ROW
EXECUTE PROCEDURE update_modified_timestamp();

-- +goose Down
DROP TRIGGER metabase_dashboard_set_modified ON metabase_dashboard;
DROP TABLE metabase_dashboard;
