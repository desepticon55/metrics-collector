package metrics

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-playground/validator/v10"
)

type Mapper struct {
	v *validator.Validate
}

func NewMapper(v *validator.Validate) Mapper {
	return Mapper{v: v}
}

func (m Mapper) MapRequestToDomainModel(dto common.MetricRequestDto) (server.Metric, error) {
	if err := m.v.Struct(dto); err != nil {
		return nil, server.NewValidationError(err)
	}

	switch dto.MType {
	case common.Gauge:
		if dto.Value == nil {
			return nil, fmt.Errorf("value is required for gauge type")
		}
		return &server.Gauge{
			BaseMetric: server.BaseMetric{Name: dto.ID, Type: common.Gauge},
			Value:      *dto.Value,
		}, nil
	case common.Counter:
		if dto.Delta == nil {
			return nil, fmt.Errorf("delta is required for counter type")
		}
		return &server.Counter{
			BaseMetric: server.BaseMetric{Name: dto.ID, Type: common.Counter},
			Value:      *dto.Delta,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", dto.MType)
	}
}

func (Mapper) MapDomainModelToResponse(domainModel server.Metric) common.MetricResponseDto {
	switch m := domainModel.(type) {
	case *server.Gauge:
		return common.MetricResponseDto{
			ID:    m.GetName(),
			MType: m.GetType(),
			Value: &m.Value,
			Delta: nil,
		}
	case *server.Counter:
		return common.MetricResponseDto{
			ID:    m.GetName(),
			MType: m.GetType(),
			Value: nil,
			Delta: &m.Value,
		}
	default:
		return common.MetricResponseDto{}
	}
}
