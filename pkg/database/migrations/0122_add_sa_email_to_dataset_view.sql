-- +goose Up
DROP VIEW IF EXISTS dataset_view CASCADE;
CREATE VIEW dataset_view AS(
    SELECT dp."group" AS dp_group,
        ds.id AS ds_id,
        ds.name AS ds_name,
        ds.description AS ds_description,
        ds.created AS ds_created,
        ds.last_modified AS ds_last_modified,
        ds.slug AS ds_slug,
        ds.pii,
        ds.keywords AS ds_keywords,
        ds.repo AS ds_repo,
        ds.target_user AS ds_target_user,
        ds.anonymisation_description AS ds_anonymisation_description,
        dsrc.id AS bq_id,
        dsrc.created AS bq_created,
        dsrc.last_modified AS bq_last_modified,
        dsrc.expires AS bq_expires,
        dsrc.description AS bq_description,
        dsrc.missing_since AS bq_missing_since,
        dsrc.pii_tags,
        dsrc.project_id AS bq_project,
        dsrc.dataset AS bq_dataset,
        dsrc.table_name AS bq_table_name,
        dsrc.table_type AS bq_table_type,
        dsrc.pseudo_columns,
        dsrc.schema AS bq_schema,
        ds.dataproduct_id AS ds_dp_id,
        omm.database_id AS omb_database_id,
        rmm.database_id AS rmb_database_id,
        rmm.sa_email AS rmb_sa_email
    FROM datasets ds
        LEFT JOIN (
            SELECT
                datasource_bigquery.dataset_id,
                datasource_bigquery.project_id,
                datasource_bigquery.dataset,
                datasource_bigquery.table_name,
                datasource_bigquery.schema,
                datasource_bigquery.last_modified,
                datasource_bigquery.created,
                datasource_bigquery.expires,
                datasource_bigquery.table_type,
                datasource_bigquery.description,
                datasource_bigquery.pii_tags,
                datasource_bigquery.missing_since,
                datasource_bigquery.id,
                datasource_bigquery.is_reference,
                datasource_bigquery.pseudo_columns,
                datasource_bigquery.deleted
            FROM datasource_bigquery
            WHERE datasource_bigquery.is_reference = false
        ) dsrc ON ds.id = dsrc.dataset_id
        LEFT JOIN dataproducts dp ON ds.dataproduct_id = dp.id
        LEFT JOIN restricted_metabase_metadata rmm ON ds.id = rmm.dataset_id
        LEFT JOIN open_metabase_metadata omm ON ds.id = omm.dataset_id
);

-- +goose Down
DROP VIEW IF EXISTS dataset_view CASCADE;
CREATE VIEW dataset_view AS(
    SELECT dp."group" AS dp_group,
        ds.id AS ds_id,
        ds.name AS ds_name,
        ds.description AS ds_description,
        ds.created AS ds_created,
        ds.last_modified AS ds_last_modified,
        ds.slug AS ds_slug,
        ds.pii,
        ds.keywords AS ds_keywords,
        ds.repo AS ds_repo,
        ds.target_user AS ds_target_user,
        ds.anonymisation_description AS ds_anonymisation_description,
        dsrc.id AS bq_id,
        dsrc.created AS bq_created,
        dsrc.last_modified AS bq_last_modified,
        dsrc.expires AS bq_expires,
        dsrc.description AS bq_description,
        dsrc.missing_since AS bq_missing_since,
        dsrc.pii_tags,
        dsrc.project_id AS bq_project,
        dsrc.dataset AS bq_dataset,
        dsrc.table_name AS bq_table_name,
        dsrc.table_type AS bq_table_type,
        dsrc.pseudo_columns,
        dsrc.schema AS bq_schema,
        ds.dataproduct_id AS ds_dp_id,
        omm.database_id AS omb_database_id,
        rmm.database_id AS rmb_database_id
    FROM datasets ds
        LEFT JOIN (
            SELECT
                datasource_bigquery.dataset_id,
                datasource_bigquery.project_id,
                datasource_bigquery.dataset,
                datasource_bigquery.table_name,
                datasource_bigquery.schema,
                datasource_bigquery.last_modified,
                datasource_bigquery.created,
                datasource_bigquery.expires,
                datasource_bigquery.table_type,
                datasource_bigquery.description,
                datasource_bigquery.pii_tags,
                datasource_bigquery.missing_since,
                datasource_bigquery.id,
                datasource_bigquery.is_reference,
                datasource_bigquery.pseudo_columns,
                datasource_bigquery.deleted
            FROM datasource_bigquery
            WHERE datasource_bigquery.is_reference = false
        ) dsrc ON ds.id = dsrc.dataset_id
        LEFT JOIN dataproducts dp ON ds.dataproduct_id = dp.id
        LEFT JOIN restricted_metabase_metadata rmm ON ds.id = rmm.dataset_id
        LEFT JOIN open_metabase_metadata omm ON ds.id = omm.dataset_id
);
