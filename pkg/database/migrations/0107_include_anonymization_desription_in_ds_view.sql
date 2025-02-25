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
        ds.target_user as ds_target_user,
        ds.anonymisation_description as ds_anonymisation_description,
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

-- +goose Down
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
        ds.target_user as ds_target_user,
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
