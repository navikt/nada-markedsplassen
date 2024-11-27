-- +goose Up
UPDATE metabase_metadata 
SET permission_group_id = '0' 
WHERE permission_group_id IS NULL AND collection_id IS NULL AND sync_completed IS NOT NULL;
