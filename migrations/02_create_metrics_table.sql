-- +goose Up
CREATE TABLE mtr_collector.metrics
(
    name       VARCHAR(50),
    type       VARCHAR(20),
    value      TEXT        NOT NULL,
    PRIMARY KEY (name, type)
);

-- +goose Down
DROP TABLE mtr_collector.metrics;