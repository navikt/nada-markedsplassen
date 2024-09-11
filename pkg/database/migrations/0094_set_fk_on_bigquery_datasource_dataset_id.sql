-- +goose Up
ALTER TABLE datasource_bigquery ADD CONSTRAINT fk_bigquery_dataset
    FOREIGN KEY (dataset_id)
        REFERENCES datasets (id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE datasource_bigquery DROP CONSTRAINT fk_bigquery_dataset;
