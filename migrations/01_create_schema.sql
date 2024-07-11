-- +goose Up
CREATE SCHEMA mtr_collector AUTHORIZATION postgres;
GRANT USAGE ON SCHEMA mtr_collector TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA mtr_collector TO postgres;

-- +goose Down
DROP SCHEMA mtr_collector;