package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Storage struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func New(pool *pgxpool.Pool, logger *zap.Logger) *Storage {
	return &Storage{
		pool:   pool,
		logger: logger,
	}
}

func (s *Storage) FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (server.Metric, bool) {
	query := "SELECT value FROM mtr_collector.metrics WHERE name=$1 AND type=$2"

	row := s.pool.QueryRow(ctx, query, metricName, metricType)

	var valueStr string
	err := row.Scan(&valueStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Info("Metric not found", zap.String("metricName", metricName), zap.String("metricType", string(metricType)))
			return nil, false
		}
		s.logger.Error("Error during find metric", zap.Error(err))
		return nil, false
	}

	var metric server.Metric
	switch metricType {
	case common.Gauge:
		metric = &server.Gauge{
			BaseMetric: server.BaseMetric{Name: metricName, Type: common.Gauge},
		}
	case common.Counter:
		metric = &server.Counter{
			BaseMetric: server.BaseMetric{Name: metricName, Type: common.Counter},
		}
	default:
		s.logger.Error("Unsupported metric type", zap.String("metricType", string(metricType)))
		return nil, false
	}

	err = metric.SetValueFromString(valueStr)
	if err != nil {
		s.logger.Error("Error setting metric value from string", zap.Error(err))
		return nil, false
	}

	return metric, true
}

func (s *Storage) FindAllMetrics(ctx context.Context) ([]server.Metric, error) {
	query := "SELECT name, type, value FROM mtr_collector.metrics"
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var metrics []server.Metric
	for rows.Next() {
		var name, metricType, valueStr string
		if err := rows.Scan(&name, &metricType, &valueStr); err != nil {
			s.logger.Error("Error scanning row", zap.Error(err))
			continue
		}

		metric, err := s.createMetricFromRow(name, metricType, valueStr)
		if err != nil {
			s.logger.Error("Error creating metric from row", zap.String("name", name), zap.String("type", metricType), zap.Error(err))
			continue
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return metrics, nil
}

func (s *Storage) SaveMetrics(ctx context.Context, metrics []server.Metric) ([]server.Metric, error) {
	savedMetrics, err := s.saveMetricsWithTx(ctx, metrics)
	if err != nil {
		s.logger.Error("Error saving metrics", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Successfully saved metrics", zap.Int("saved_metrics_count", len(savedMetrics)))
	return savedMetrics, nil
}

func (s *Storage) saveMetricsWithTx(ctx context.Context, metrics []server.Metric) ([]server.Metric, error) {
	var savedMetrics []server.Metric
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	for _, metric := range metrics {
		savedMetric, e := s.saveMetricWithTx(ctx, tx, metric)
		if e != nil {
			rollbackErr := tx.Rollback(ctx)
			return nil, errors.Join(e, rollbackErr)
		}
		savedMetrics = append(savedMetrics, savedMetric)
	}
	err = tx.Commit(ctx)
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		return nil, errors.Join(err, rollbackErr)
	}
	return savedMetrics, nil
}

func (s *Storage) saveMetricWithTx(ctx context.Context, tx pgx.Tx, metric server.Metric) (server.Metric, error) {
	valueStr, err := metric.GetValueAsString()
	if err != nil {
		return nil, err
	}

	query := `
        INSERT INTO mtr_collector.metrics (name, type, value)
        VALUES ($1, $2, $3)
        ON CONFLICT (name, type)
        DO UPDATE SET value = 
        CASE
            WHEN metrics.type = 'counter' THEN 
                (CAST(metrics.value AS BIGINT) + CAST(EXCLUDED.value AS BIGINT))::TEXT
            ELSE EXCLUDED.value
        END
        RETURNING value
    `
	var updatedValueStr string
	err = tx.QueryRow(ctx, query, metric.GetName(), metric.GetType(), valueStr).Scan(&updatedValueStr)
	if err != nil {
		return nil, err
	}

	err = metric.SetValueFromString(updatedValueStr)
	if err != nil {
		return nil, err
	}

	return metric, nil
}

func (s *Storage) createMetricFromRow(name, metricType, valueStr string) (server.Metric, error) {
	var metric server.Metric
	switch common.MetricType(metricType) {
	case common.Gauge:
		metric = &server.Gauge{
			BaseMetric: server.BaseMetric{Name: name, Type: common.Gauge},
		}
	case common.Counter:
		metric = &server.Counter{
			BaseMetric: server.BaseMetric{Name: name, Type: common.Counter},
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metricType)
	}

	err := metric.SetValueFromString(valueStr)
	if err != nil {
		return nil, err
	}

	return metric, nil
}
