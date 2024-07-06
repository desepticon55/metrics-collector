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

func (s *Storage) SaveMetric(ctx context.Context, metric server.Metric) (server.Metric, error) {
	var foundMetric server.Metric
	var exists bool

	query := "SELECT value FROM mt_cl.metrics WHERE name=$1 AND type=$2"
	row := s.connection.QueryRow(ctx, query, metric.GetName(), metric.GetType())
	var valueStr string
	err := row.Scan(&valueStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			exists = false
		} else {
			return nil, err
		}
	} else {
		exists = true
		switch metric.GetType() {
		case common.Gauge:
			foundMetric = &server.Gauge{
				BaseMetric: server.BaseMetric{Name: metric.GetName(), Type: common.Gauge},
			}
		case common.Counter:
			foundMetric = &server.Counter{
				BaseMetric: server.BaseMetric{Name: metric.GetName(), Type: common.Counter},
			}
		default:
			return nil, errors.New("unsupported metric type")
		}

		err = foundMetric.SetValueFromString(valueStr)
		if err != nil {
			return nil, err
		}
	}

	valueStr, err = metric.GetValueAsString()
	if err != nil {
		return nil, err
	}

	if !exists {
		query = "INSERT INTO mt_cl.metrics (name, type, value) VALUES ($1, $2, $3)"
		_, err := s.connection.Exec(ctx, query, metric.GetName(), metric.GetType(), valueStr)
		if err != nil {
			return nil, err
		}
	} else {
		if metric.GetType() == common.Counter {
			oldValueStr, err := foundMetric.GetValueAsString()
			if err != nil {
				return nil, err
			}
			oldValue, err := strconv.ParseInt(oldValueStr, 10, 64)
			if err != nil {
				return nil, err
			}
			newValue, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				return nil, err
			}
			metric.(*server.Counter).Value = oldValue + newValue
			valueStr, err = metric.GetValueAsString()
			if err != nil {
				return nil, err
			}
		}
		query = "UPDATE mt_cl.metrics SET value=$1 WHERE name=$2 AND type=$3"
		_, err := s.connection.Exec(ctx, query, valueStr, metric.GetName(), metric.GetType())
		if err != nil {
			return nil, err
		}
	}

	return metric, nil
}

func (s *Storage) FindOneMetric(ctx context.Context, metricName string, metricType common.MetricType) (server.Metric, bool) {
	query := "SELECT name, type, value FROM mt_cl.metrics WHERE name=$1 AND type=$2"
	row := s.connection.QueryRow(ctx, query, metricName, metricType)

	var name string
	var metricTypeStr string
	var valueStr string

	err := row.Scan(&name, &metricTypeStr, &valueStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		s.logger.Error("Error during find metric: ", zap.Error(err))
		return nil, false
	}

	var metric server.Metric
	switch common.MetricType(metricTypeStr) {
	case common.Gauge:
		metric = &server.Gauge{
			BaseMetric: server.BaseMetric{Name: name, Type: common.Gauge},
		}
	case common.Counter:
		metric = &server.Counter{
			BaseMetric: server.BaseMetric{Name: name, Type: common.Counter},
		}
	default:
		s.logger.Error(fmt.Sprintf("Unsupported metric type: %s", metricTypeStr))
		return nil, false
	}

	err = metric.SetValueFromString(valueStr)
	if err != nil {
		s.logger.Error("Error during set value from string", zap.Error(err))
		return nil, false
	}

	return metric, true
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
			s.logger.Error(fmt.Sprintf("Unsupported metric type: %s", metricType))
			continue
		}

		err = metric.SetValueFromString(valueStr)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}
