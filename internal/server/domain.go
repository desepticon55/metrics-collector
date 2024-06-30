package server

import (
	"encoding/json"
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
)

type Metric struct {
	Name      string            `json:"name"`
	Type      common.MetricType `json:"type"`
	Value     interface{}       `json:"value"`
	ValueType string            `json:"valueType"`
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	var temp struct {
		Name      string            `json:"name"`
		Type      common.MetricType `json:"type"`
		ValueType string            `json:"valueType"`
		Value     interface{}       `json:"value"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	m.Name = temp.Name
	m.Type = temp.Type
	m.ValueType = temp.ValueType
	switch temp.ValueType {
	case "float64":
		m.Value = temp.Value.(float64)
	case "int64":
		if val, ok := temp.Value.(float64); ok {
			m.Value = int64(val)
		}
	default:
		return fmt.Errorf("unsupported value type: %s", temp.ValueType)
	}
	return nil
}
