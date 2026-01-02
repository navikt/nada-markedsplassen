-- +goose Up
ALTER TABLE dataset_access ADD COLUMN platform TEXT NOT NULL DEFAULT 'bigquery';
ALTER TABLE dataset_access ALTER COLUMN platform drop DEFAULT;

ALTER TABLE dataset_access_requests ADD COLUMN platform TEXT NOT NULL DEFAULT 'bigquery';
ALTER TABLE dataset_access_requests ALTER COLUMN platform drop DEFAULT;

INSERT INTO dataset_access (dataset_id, subject, granter, expires, created, revoked, access_request_id, owner, platform)
SELECT da.dataset_id, da.subject, da.granter, da.expires, da.created, da.revoked, da.access_request_id, da.owner, 'metabase'
FROM dataset_access da
    JOIN open_metabase_metadata omm ON da.dataset_id = omm.dataset_id
UNION
SELECT da.dataset_id, da.subject, da.granter, da.expires, da.created, da.revoked, da.access_request_id, da.owner, 'metabase'
FROM dataset_access da
    JOIN restricted_metabase_metadata rmm ON da.dataset_id = rmm.dataset_id;

DROP VIEW dataset_access_view;
CREATE VIEW dataset_access_view AS (
    SELECT 
        da.id as access_id,
        da.subject as access_subject,
        da.owner as access_owner,
        da.granter as access_granter,
        da.expires as access_expires,
        da.created as access_created,
        da.revoked as access_revoked,
        da.dataset_id as access_dataset_id,
        da.access_request_id as access_request_id,
        da.platform as access_platform,
        dar.owner AS access_request_owner,
        dar.subject AS access_request_subject,
        dar.last_modified AS access_request_last_modified,
        dar.created AS access_request_created,
        dar.expires AS access_request_expires,
        dar.status AS access_request_status,
        dar.closed AS access_request_closed,
        dar.granter AS access_request_granter,
        dar.reason AS access_request_reason,
        pdoc.id AS polly_id,
        pdoc.name AS polly_name,
        pdoc.url AS polly_url,
        pdoc.external_id AS polly_external_id
    FROM dataset_access AS da
    LEFT JOIN dataset_access_requests AS dar ON dar.id = da.access_request_id
    LEFT JOIN polly_documentation AS pdoc ON pdoc.id = dar.polly_documentation_id
);

-- +goose Down

DELETE
FROM dataset_access
WHERE platform != 'bigquery';

DROP VIEW dataset_access_view;
CREATE VIEW dataset_access_view AS (
    SELECT 
        da.id as access_id,
        da.subject as access_subject,
        da.owner as access_owner,
        da.granter as access_granter,
        da.expires as access_expires,
        da.created as access_created,
        da.revoked as access_revoked,
        da.dataset_id as access_dataset_id,
        da.access_request_id as access_request_id,
        dar.owner AS access_request_owner,
        dar.subject AS access_request_subject,
        dar.last_modified AS access_request_last_modified,
        dar.created AS access_request_created,
        dar.expires AS access_request_expires,
        dar.status AS access_request_status,
        dar.closed AS access_request_closed,
        dar.granter AS access_request_granter,
        dar.reason AS access_request_reason,
        pdoc.id AS polly_id,
        pdoc.name AS polly_name,
        pdoc.url AS polly_url,
        pdoc.external_id AS polly_external_id
    FROM dataset_access AS da
    LEFT JOIN dataset_access_requests AS dar ON dar.id = da.access_request_id
    LEFT JOIN polly_documentation AS pdoc ON pdoc.id = dar.polly_documentation_id
);

ALTER TABLE dataset_access DROP COLUMN platform;
ALTER TABLE dataset_access_requests DROP COLUMN platform;
