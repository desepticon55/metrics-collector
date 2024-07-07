package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"strconv"
)

type Storage struct {
	connection *pgx.Conn
	logger     *zap.Logger
}

func New(connection *pgx.Conn, logger *zap.Logger) *Storage {
	return &Storage{
		connection: connection,
		logger:     logger,
	}
}

func (s *Storage) SaveMetrics(ctx context.Context, metrics []server.Metric) ([]server.Metric, error) {
	tx, err := s.connection.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var savedMetrics []server.Metric

	for _, metric := range metrics {
		savedMetric, err := s.saveMetricWithTx(ctx, tx, metric)
		if err != nil {
			return savedMetrics, err
		}
		savedMetrics = append(savedMetrics, savedMetric)
	}

	return savedMetrics, nil
}

func (s *Storage) FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (server.Metric, bool) {
	tx, err := s.connection.Begin(ctx)
	if err != nil {
		s.logger.Error("Error during begin transaction", zap.Error(err))
		return nil, false
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	metric, exists, err := s.findOneMetricWithTx(ctx, tx, metricName, metricType)
	if err != nil {
		s.logger.Error("Error during find metric", zap.Error(err))
		return nil, false
	}

	return metric, exists
}

func (s *Storage) FindAllMetrics(ctx context.Context) ([]server.Metric, error) {
	query := "SELECT name, type, value FROM mt_cl.metrics"
	rows, err := s.connection.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []server.Metric
	for rows.Next() {
		var name string
		var metricType string
		var valueStr string

		err := rows.Scan(&name, &metricType, &valueStr)
		if err != nil {
			return nil, err
		}

		metric, err := s.createMetricFromRow(name, metricType, valueStr)
		if err != nil {
			s.logger.Error("Error during create metric from row", zap.Error(err))
			continue
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *Storage) findOneMetricWithTx(ctx context.Context, tx pgx.Tx, metricName string, metricType common.MetricType) (server.Metric, bool, error) {
	query := "SELECT value FROM mt_cl.metrics WHERE name=$1 AND type=$2"
	row := tx.QueryRow(ctx, query, metricName, metricType)

	var valueStr string
	err := row.Scan(&valueStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
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
		return nil, false, errors.New("unsupported metric type")
	}

	err = metric.SetValueFromString(valueStr)
	if err != nil {
		return nil, false, err
	}

	return metric, true, nil
}

func (s *Storage) saveMetricWithTx(ctx context.Context, tx pgx.Tx, metric server.Metric) (server.Metric, error) {
	existingMetric, exists, err := s.findOneMetricWithTx(ctx, tx, metric.GetName(), metric.GetType())
	if err != nil {
		return nil, err
	}

	if !exists {
		err = s.createMetric(ctx, tx, metric)
		if err != nil {
			return nil, err
		}
	} else {
		if metric.GetType() == common.Counter {
			err = s.updateCounterValues(existingMetric, metric)
			if err != nil {
				return nil, err
			}
		}
		valueStr, err := metric.GetValueAsString()
		if err != nil {
			return nil, err
		}
		query := "UPDATE mt_cl.metrics SET value=$1 WHERE name=$2 AND type=$3"
		_, err = tx.Exec(ctx, query, valueStr, metric.GetName(), metric.GetType())
		if err != nil {
			return nil, err
		}
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

func (s *Storage) createMetric(ctx context.Context, tx pgx.Tx, metric server.Metric) error {
	valueStr, err := metric.GetValueAsString()
	if err != nil {
		return err
	}

	query := "INSERT INTO mt_cl.metrics (name, type, value) VALUES ($1, $2, $3)"
	_, err = tx.Exec(ctx, query, metric.GetName(), metric.GetType(), valueStr)
	return err
}

func (s *Storage) updateCounterValues(existingMetric, newMetric server.Metric) error {
	oldValueStr, err := existingMetric.GetValueAsString()
	if err != nil {
		return err
	}
	oldValue, err := strconv.ParseInt(oldValueStr, 10, 64)
	if err != nil {
		return err
	}
	newValueStr, err := newMetric.GetValueAsString()
	if err != nil {
		return err
	}
	newValue, err := strconv.ParseInt(newValueStr, 10, 64)
	if err != nil {
		return err
	}
	newMetric.(*server.Counter).Value = oldValue + newValue
	return nil
}
