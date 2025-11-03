-- +goose Up
CREATE TABLE open_metabase_metadata (
    "dataset_id" UUID PRIMARY KEY,
    "database_id" INT,
    "deleted_at" TIMESTAMPTZ,
    "sync_completed" TIMESTAMPTZ,
    CONSTRAINT fk_open_metabase_metadata
    FOREIGN KEY (dataset_id)
        REFERENCES datasets (id) ON DELETE CASCADE);

INSERT INTO open_metabase_metadata
(SELECT dataset_id, database_id, deleted_at, sync_completed
FROM metabase_metadata WHERE permission_group_id = 0);

DELETE FROM metabase_metadata WHERE permission_group_id = 0;

ALTER TABLE metabase_metadata RENAME TO restricted_metabase_metadata;

DROP VIEW dataset_view;
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
        omm.deleted_at AS omb_deleted_at,
        rmm.database_id AS rmb_database_id,
        rmm.deleted_at AS rmb_deleted_at
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

INSERT INTO restricted_metabase_metadata
(SELECT database_id, 0, '', NULL, deleted_at, dataset_id, sync_completed, NULL
FROM open_metabase_metadata);

UPDATE restricted_metabase_metadata
SET permission_group_id = 0
WHERE permission_group_id IS NULL; 

ALTER TABLE restricted_metabase_metadata RENAME TO metabase_metadata;

DROP VIEW dataset_view;
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
        mm.database_id AS mb_database_id,
        mm.deleted_at AS mb_deleted_at
       FROM (((datasets ds
         LEFT JOIN ( SELECT datasource_bigquery.dataset_id,
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
              WHERE (datasource_bigquery.is_reference = false)) dsrc ON ((ds.id = dsrc.dataset_id)))
         LEFT JOIN dataproducts dp ON ((ds.dataproduct_id = dp.id)))
         LEFT JOIN metabase_metadata mm ON ((ds.id = mm.dataset_id)))
);

DROP TABLE open_metabase_metadata;