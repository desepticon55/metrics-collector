package server

type MetricType string

type Metric struct {
	Name  string
	Type  MetricType
	Value interface{}
}

type MetricRequestDto struct {
	Name  string     `json:"name" validate:"required"`
	Type  MetricType `json:"type" validate:"metricTypeCheck"`
	Value string     `json:"value" validate:"required"`
}

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
