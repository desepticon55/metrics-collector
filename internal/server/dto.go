package server

type MetricType string

type MetricRequestDto struct {
	Name  string     `json:"name" validate:"required"`
	Type  MetricType `json:"type" validate:"required"`
	Value string     `json:"value" validate:"required"`
}

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)
