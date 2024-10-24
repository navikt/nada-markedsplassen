-- +goose Up
DROP VIEW dataset_view;
CREATE VIEW dataset_view AS(
    SELECT
        dp.group as dp_group,
        ds.id as ds_id,
        ds.name as ds_name,
        ds.description as ds_description,
        ds.created as ds_created,
        ds.last_modified as ds_last_modified,
        ds.slug as ds_slug,
        ds.pii as pii,
        ds.keywords as ds_keywords,
        ds.repo as ds_repo,
        dsrc.id AS bq_id,
        dsrc.created as bq_created,
        dsrc.last_modified as bq_last_modified,
        dsrc.expires as bq_expires,
        dsrc.description as bq_description,
        dsrc.missing_since as bq_missing_since,
        dsrc.pii_tags as pii_tags,
        dsrc.project_id as bq_project,
        dsrc.dataset as bq_dataset,
        dsrc.table_name as bq_table_name,
        dsrc.table_type as bq_table_type,
        dsrc.pseudo_columns as pseudo_columns,
        dsrc.schema as bq_schema,
        ds.dataproduct_id as ds_dp_id,
        dm.services as mapping_services,
        mm.database_id as mb_database_id,
        mm.deleted_at as mb_deleted_at
    FROM
        datasets ds
        LEFT JOIN (
            SELECT
                *
            FROM
                datasource_bigquery
            WHERE
                is_reference = false
        ) dsrc ON ds.id = dsrc.dataset_id
        LEFT JOIN third_party_mappings dm ON ds.id = dm.dataset_id
        LEFT JOIN dataproducts dp ON ds.dataproduct_id = dp.id
        LEFT JOIN metabase_metadata mm ON ds.id = mm.dataset_id
);

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

-- +goose Down
DROP VIEW dataset_access_view;
CREATE VIEW dataset_access_view AS (
    SELECT da.*,
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

DROP VIEW dataset_view;
CREATE VIEW dataset_view AS(
    SELECT
        ds.id as ds_id,
        ds.name as ds_name,
        ds.description as ds_description,
        ds.created as ds_created,
        ds.last_modified as ds_last_modified,
        ds.slug as ds_slug,
        ds.pii as pii,
        ds.keywords as ds_keywords,
        ds.repo as ds_repo,
        dsrc.id AS bq_id,
        dsrc.created as bq_created,
        dsrc.last_modified as bq_last_modified,
        dsrc.expires as bq_expires,
        dsrc.description as bq_description,
        dsrc.missing_since as bq_missing_since,
        dsrc.pii_tags as pii_tags,
        dsrc.project_id as bq_project,
        dsrc.dataset as bq_dataset,
        dsrc.table_name as bq_table_name,
        dsrc.table_type as bq_table_type,
        dsrc.pseudo_columns as pseudo_columns,
        dsrc.schema as bq_schema,
        ds.dataproduct_id as ds_dp_id,
        dm.services as mapping_services,
        da.id as access_id,
        da.subject as access_subject,
        da.owner as access_owner,
        da.granter as access_granter,
        da.expires as access_expires,
        da.created as access_created,
        da.revoked as access_revoked,
        da.access_request_owner AS access_request_owner,
        da.access_request_subject AS access_request_subject,
        da.access_request_last_modified AS access_request_last_modified,
        da.access_request_created AS access_request_created,
        da.access_request_expires AS access_request_expires,
        da.access_request_status AS access_request_status,
        da.access_request_closed AS access_request_closed,
        da.access_request_granter AS access_request_granter,
        da.access_request_reason AS access_request_reason,
        da.polly_id AS polly_id,
        da.polly_name AS polly_name,
        da.polly_url AS polly_url,
        da.polly_external_id AS polly_external_id,
        da.access_request_id as access_request_id,
        mm.database_id as mb_database_id,
        mm.deleted_at as mb_deleted_at
    FROM
        datasets ds
        LEFT JOIN (
            SELECT
                *
            FROM
                datasource_bigquery
            WHERE
                is_reference = false
        ) dsrc ON ds.id = dsrc.dataset_id
        LEFT JOIN third_party_mappings dm ON ds.id = dm.dataset_id
        LEFT JOIN dataset_access_view da ON  ds.id = da.dataset_id
        LEFT JOIN metabase_metadata mm ON ds.id = mm.dataset_id
);
