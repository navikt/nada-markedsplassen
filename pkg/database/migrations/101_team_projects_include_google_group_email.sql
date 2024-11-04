-- +goose Up
ALTER TABLE team_projects ADD COLUMN "group_email" TEXT;
UPDATE team_projects tp SET group_email = CONCAT(tp.team,'@nav.no');
ALTER TABLE team_projects ALTER COLUMN "group_email" SET NOT NULL;

-- +goose Down
ALTER TABLE team_projects DROP COLUMN "group_email";
