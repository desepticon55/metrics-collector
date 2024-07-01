package metrics

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-playground/validator/v10"
)

type mapper struct {
}

func NewMapper() mapper {
	return mapper{}
}

func (mapper) MapRequestToDomainModel(dto common.MetricRequestDto) (server.Metric, error) {
	validate := validator.New()
	validate.RegisterStructValidation(server.MetricValidator, common.MetricRequestDto{})
	if err := validate.Struct(dto); err != nil {
		return server.Metric{}, server.NewValidationError(err)
	}

	switch dto.MType {
	case common.Gauge:
		return server.Metric{Name: dto.ID, Type: common.Gauge, Value: *dto.Value, ValueType: "float64"}, nil
	case common.Counter:
		return server.Metric{Name: dto.ID, Type: common.Counter, Value: *dto.Delta, ValueType: "int64"}, nil
	default:
		return server.Metric{}, fmt.Errorf("unsupported metric type: %s", dto.MType)
	}
}

func (mapper) MapDomainModelToResponse(domainModel server.Metric) common.MetricResponseDto {
	if domainModel.Type == common.Gauge {
		value := domainModel.Value.(float64)
		return common.MetricResponseDto{
			ID:    domainModel.Name,
			MType: domainModel.Type,
			Value: &value,
			Delta: nil,
		}
	}
	delta := domainModel.Value.(int64)
	return common.MetricResponseDto{
		ID:    domainModel.Name,
		MType: domainModel.Type,
		Value: nil,
		Delta: &delta,
	}
}
