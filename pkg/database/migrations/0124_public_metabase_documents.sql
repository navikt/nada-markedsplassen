-- +goose Up
CREATE TABLE metabase_document (
    "id"                 uuid        DEFAULT uuid_generate_v4(),
    "name"               TEXT        NOT NULL,
    "description"        TEXT,
    "group"              TEXT        NOT NULL,
    "public_document_id" uuid        NOT NULL,
    "metabase_id"        INT         NOT NULL UNIQUE CHECK ("metabase_id" > 0),
    "created_by"         TEXT        NOT NULL,
    "created"            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "last_modified"      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "tsv_document"       tsvector GENERATED ALWAYS AS (
                                 to_tsvector('norwegian', "name")
                                 || to_tsvector('norwegian', coalesce("description", ''))
                         ) STORED,
    "keywords"           TEXT[]      NOT NULL DEFAULT '{}',
    "teamkatalogen_url"  TEXT,
    "team_id"            uuid,
    PRIMARY KEY (id)
);

CREATE INDEX metabase_document_group_idx ON metabase_document ("group");

CREATE TRIGGER metabase_document_set_modified
    BEFORE UPDATE
    ON metabase_document
    FOR EACH ROW
EXECUTE PROCEDURE update_modified_timestamp();

-- +goose Down
DROP TRIGGER metabase_document_set_modified ON metabase_document;
DROP INDEX metabase_document_group_idx;
DROP TABLE metabase_document;
