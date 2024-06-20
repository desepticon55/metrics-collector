package metrics

import (
	"fmt"
	"github.com/desepticon55/metrics-collector/internal/common"
	"github.com/desepticon55/metrics-collector/internal/server"
	"github.com/go-playground/validator/v10"
	"strconv"
	"strings"
)

type mapper struct {
}

func NewMapper() mapper {
	return mapper{}
}

func (mapper) MapRequestToDomainModel(dto common.MetricRequestDto) (common.Metric, error) {
	validate := validator.New()
	validate.RegisterValidation("metricTypeCheck", server.MetricTypeValidator)
	if err := validate.Struct(dto); err != nil {
		return common.Metric{}, server.NewValidationError(err)
	}

	switch dto.Type {
	case common.Gauge:
		value, err := strconv.ParseFloat(strings.TrimSpace(dto.Value), 64)
		if err != nil {
			return common.Metric{}, server.NewIncorrectMetricValueError(dto.Name, dto.Type, dto.Value, "float64")
		}
		return common.Metric{Name: dto.Name, Type: common.Gauge, Value: value}, nil
	case common.Counter:
		value, err := strconv.ParseInt(strings.TrimSpace(dto.Value), 10, 64)
		if err != nil {
			return common.Metric{}, server.NewIncorrectMetricValueError(dto.Name, dto.Type, dto.Value, "int64")
		}
		return common.Metric{Name: dto.Name, Type: common.Counter, Value: value}, nil
	default:
		return common.Metric{}, fmt.Errorf("unsupported metric type: %s", dto.Type)
	}
}
