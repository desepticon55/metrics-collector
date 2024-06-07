package common

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

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

type MetricResponseDto struct {
	Name  string
	Type  MetricType
	Value string
}
