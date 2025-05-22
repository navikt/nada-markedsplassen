-- +goose Up
DROP VIEW IF EXISTS search CASCADE;
DROP VIEW IF EXISTS dataset_view CASCADE;
DROP TABLE third_party_mappings CASCADE;

CREATE VIEW search AS (
SELECT dp.id                              AS element_id,
       'dataproduct'::text                AS element_type,
       COALESCE(dp.description, ''::text) AS description,
       dpk.aggregated_keywords            AS keywords,
       dp."group",
       dp.team_id,
       dp.created,
       dp.last_modified,
       (setweight(to_tsvector('norwegian'::regconfig, dp.name), 'A'::"char") ||
        setweight(to_tsvector('norwegian'::regconfig, COALESCE(dp.description, ''::text)), 'B'::"char")) ||
       setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(dp."group", ''::text), '@'::text, 1)),
                 'D'::"char")             AS tsv_document
FROM dataproducts dp
         LEFT JOIN (SELECT dk.dataproduct_id,
                           COALESCE(array_agg(dk.flatterned_keywords_array), '{}'::text[]) AS aggregated_keywords
                    FROM (SELECT datasets.dataproduct_id,
                                 unnest(datasets.keywords) AS flatterned_keywords_array
                          FROM datasets) dk
                    GROUP BY dk.dataproduct_id) dpk ON dp.id = dpk.dataproduct_id
UNION
SELECT ds.id                              AS element_id,
       'dataset'::text                    AS element_type,
       COALESCE(ds.description, ''::text) AS description,
       ds.keywords,
       dp."group",
       dp.team_id,
       ds.created,
       ds.last_modified,
       ((((setweight(to_tsvector('norwegian'::regconfig, ds.name), 'A'::"char") ||
           setweight(to_tsvector('norwegian'::regconfig, COALESCE(ds.description, ''::text)), 'B'::"char")) ||
          setweight(to_tsvector('norwegian'::regconfig, COALESCE(f_arr2text(ds.keywords), ''::text)), 'C'::"char")) ||
         setweight(to_tsvector('norwegian'::regconfig, COALESCE(ds.repo, ''::text)), 'D'::"char")) ||
        setweight(to_tsvector('norwegian'::regconfig, ds.type::text), 'D'::"char")) ||
       setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(dp."group", ''::text), '@'::text, 1)),
                 'D'::"char")             AS tsv_document
FROM datasets ds
         JOIN dataproducts dp ON ds.dataproduct_id = dp.id
UNION
SELECT ss.id                  AS element_id,
       'story'::text          AS element_type,
       ss.description,
       ss.keywords,
       ss."group",
       ss.team_id,
       ss.created,
       ss.last_modified,
       (((setweight(to_tsvector('norwegian'::regconfig, ss.name), 'A'::"char") ||
          setweight(to_tsvector('norwegian'::regconfig, ss.description), 'B'::"char")) ||
         setweight(to_tsvector('norwegian'::regconfig, COALESCE(f_arr2text(ss.keywords), ''::text)), 'C'::"char")) ||
        setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(ss.creator, ''::text), '@'::text, 1)),
                  'D'::"char")) ||
       setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(ss."group", ''::text), '@'::text, 1)),
                 'D'::"char") AS tsv_document
FROM stories ss);

CREATE VIEW dataset_view AS
SELECT dp."group"                   AS dp_group,
       ds.id                        AS ds_id,
       ds.name                      AS ds_name,
       ds.description               AS ds_description,
       ds.created                   AS ds_created,
       ds.last_modified             AS ds_last_modified,
       ds.slug                      AS ds_slug,
       ds.pii,
       ds.keywords                  AS ds_keywords,
       ds.repo                      AS ds_repo,
       ds.target_user               AS ds_target_user,
       ds.anonymisation_description AS ds_anonymisation_description,
       dsrc.id                      AS bq_id,
       dsrc.created                 AS bq_created,
       dsrc.last_modified           AS bq_last_modified,
       dsrc.expires                 AS bq_expires,
       dsrc.description             AS bq_description,
       dsrc.missing_since           AS bq_missing_since,
       dsrc.pii_tags,
       dsrc.project_id              AS bq_project,
       dsrc.dataset                 AS bq_dataset,
       dsrc.table_name              AS bq_table_name,
       dsrc.table_type              AS bq_table_type,
       dsrc.pseudo_columns,
       dsrc.schema                  AS bq_schema,
       ds.dataproduct_id            AS ds_dp_id,
       mm.database_id               AS mb_database_id,
       mm.deleted_at                AS mb_deleted_at
FROM datasets ds
         LEFT JOIN (SELECT datasource_bigquery.dataset_id,
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
                    WHERE datasource_bigquery.is_reference = false) dsrc ON ds.id = dsrc.dataset_id
         LEFT JOIN dataproducts dp ON ds.dataproduct_id = dp.id
         LEFT JOIN metabase_metadata mm ON ds.id = mm.dataset_id;


-- +goose Down
DROP VIEW IF EXISTS search CASCADE;
DROP VIEW IF EXISTS dataset_view CASCADE;

create table third_party_mappings
(
    services   text[] not null,
    dataset_id uuid   not null
        primary key
        constraint fk_tpm_dataset
            references datasets
            on delete cascade
);

CREATE VIEW dataset_view AS
SELECT dp."group"                   AS dp_group,
       ds.id                        AS ds_id,
       ds.name                      AS ds_name,
       ds.description               AS ds_description,
       ds.created                   AS ds_created,
       ds.last_modified             AS ds_last_modified,
       ds.slug                      AS ds_slug,
       ds.pii,
       ds.keywords                  AS ds_keywords,
       ds.repo                      AS ds_repo,
       ds.target_user               AS ds_target_user,
       ds.anonymisation_description AS ds_anonymisation_description,
       dsrc.id                      AS bq_id,
       dsrc.created                 AS bq_created,
       dsrc.last_modified           AS bq_last_modified,
       dsrc.expires                 AS bq_expires,
       dsrc.description             AS bq_description,
       dsrc.missing_since           AS bq_missing_since,
       dsrc.pii_tags,
       dsrc.project_id              AS bq_project,
       dsrc.dataset                 AS bq_dataset,
       dsrc.table_name              AS bq_table_name,
       dsrc.table_type              AS bq_table_type,
       dsrc.pseudo_columns,
       dsrc.schema                  AS bq_schema,
       ds.dataproduct_id            AS ds_dp_id,
       dm.services                  AS mapping_services,
       mm.database_id               AS mb_database_id,
       mm.deleted_at                AS mb_deleted_at
FROM datasets ds
         LEFT JOIN (SELECT datasource_bigquery.dataset_id,
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
                    WHERE datasource_bigquery.is_reference = false) dsrc ON ds.id = dsrc.dataset_id
         LEFT JOIN third_party_mappings dm ON ds.id = dm.dataset_id
         LEFT JOIN dataproducts dp ON ds.dataproduct_id = dp.id
         LEFT JOIN metabase_metadata mm ON ds.id = mm.dataset_id;

CREATE VIEW search AS
SELECT dp.id                              AS element_id,
       'dataproduct'::text                AS element_type,
       COALESCE(dp.description, ''::text) AS description,
       dpk.aggregated_keywords            AS keywords,
       dp."group",
       dp.team_id,
       dp.created,
       dp.last_modified,
       (setweight(to_tsvector('norwegian'::regconfig, dp.name), 'A'::"char") ||
        setweight(to_tsvector('norwegian'::regconfig, COALESCE(dp.description, ''::text)), 'B'::"char")) ||
       setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(dp."group", ''::text), '@'::text, 1)),
                 'D'::"char")             AS tsv_document,
       '{}'::text[]                       AS services
FROM dataproducts dp
         LEFT JOIN (SELECT dk.dataproduct_id,
                           COALESCE(array_agg(dk.flatterned_keywords_array), '{}'::text[]) AS aggregated_keywords
                    FROM (SELECT datasets.dataproduct_id,
                                 unnest(datasets.keywords) AS flatterned_keywords_array
                          FROM datasets) dk
                    GROUP BY dk.dataproduct_id) dpk ON dp.id = dpk.dataproduct_id
UNION
SELECT ds.id                              AS element_id,
       'dataset'::text                    AS element_type,
       COALESCE(ds.description, ''::text) AS description,
       ds.keywords,
       dp."group",
       dp.team_id,
       ds.created,
       ds.last_modified,
       ((((setweight(to_tsvector('norwegian'::regconfig, ds.name), 'A'::"char") ||
           setweight(to_tsvector('norwegian'::regconfig, COALESCE(ds.description, ''::text)), 'B'::"char")) ||
          setweight(to_tsvector('norwegian'::regconfig, COALESCE(f_arr2text(ds.keywords), ''::text)), 'C'::"char")) ||
         setweight(to_tsvector('norwegian'::regconfig, COALESCE(ds.repo, ''::text)), 'D'::"char")) ||
        setweight(to_tsvector('norwegian'::regconfig, ds.type::text), 'D'::"char")) ||
       setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(dp."group", ''::text), '@'::text, 1)),
                 'D'::"char")             AS tsv_document,
       tpm.services
FROM datasets ds
         JOIN dataproducts dp ON ds.dataproduct_id = dp.id
         LEFT JOIN third_party_mappings tpm ON tpm.dataset_id = ds.id
UNION
SELECT ss.id                  AS element_id,
       'story'::text          AS element_type,
       ss.description,
       ss.keywords,
       ss."group",
       ss.team_id,
       ss.created,
       ss.last_modified,
       (((setweight(to_tsvector('norwegian'::regconfig, ss.name), 'A'::"char") ||
          setweight(to_tsvector('norwegian'::regconfig, ss.description), 'B'::"char")) ||
         setweight(to_tsvector('norwegian'::regconfig, COALESCE(f_arr2text(ss.keywords), ''::text)), 'C'::"char")) ||
        setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(ss.creator, ''::text), '@'::text, 1)),
                  'D'::"char")) ||
       setweight(to_tsvector('norwegian'::regconfig, split_part(COALESCE(ss."group", ''::text), '@'::text, 1)),
                 'D'::"char") AS tsv_document,
       '{}'::text[]           AS services
FROM stories ss;
