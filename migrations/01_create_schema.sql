-- +goose Up
CREATE SCHEMA mt_cl AUTHORIZATION postgres;
GRANT USAGE ON SCHEMA mt_cl TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA mt_cl TO postgres;

-- +goose Down
DROP SCHEMA mt_cl;