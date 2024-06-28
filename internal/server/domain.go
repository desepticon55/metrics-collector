package server

import "github.com/desepticon55/metrics-collector/internal/common"

type Metric struct {
	Name  string
	Type  common.MetricType
	Value interface{}
}
