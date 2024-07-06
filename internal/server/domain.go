package server

import (
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"strconv"
)

type Metric interface {
	GetName() string
	GetType() common.MetricType
	GetValueAsString() (string, error)
	SetValueFromString(string) error
}

type BaseMetric struct {
	Name string            `json:"name"`
	Type common.MetricType `json:"type"`
}

func (m BaseMetric) GetName() string {
	return m.Name
}

func (m BaseMetric) GetType() common.MetricType {
	return m.Type
}

type Gauge struct {
	BaseMetric
	Value float64 `json:"value"`
}

func (g *Gauge) GetValueAsString() (string, error) {
	return strconv.FormatFloat(g.Value, 'f', -1, 64), nil
}

func (g *Gauge) SetValueFromString(valueStr string) error {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return err
	}
	g.Value = value
	return nil
}

type Counter struct {
	BaseMetric
	Value int64 `json:"value"`
}

func (c *Counter) GetValueAsString() (string, error) {
	return strconv.FormatInt(c.Value, 10), nil
}

func (c *Counter) SetValueFromString(valueStr string) error {
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return err
	}
	c.Value = value
	return nil
}

func MarshalMetric(metric Metric) ([]byte, error) {
	return json.Marshal(metric)
}

func UnmarshalMetric(data []byte) (Metric, error) {
	var base BaseMetric
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Type {
	case common.Gauge:
		var gauge Gauge
		if err := json.Unmarshal(data, &gauge); err != nil {
			return nil, err
		}
		return &gauge, nil
	case common.Counter:
		var counter Counter
		if err := json.Unmarshal(data, &counter); err != nil {
			return nil, err
		}
		return &counter, nil
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", base.Type)
	}
}
