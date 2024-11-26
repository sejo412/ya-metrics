package server

import (
	"fmt"
	. "github.com/sejo412/ya-metrics/internal/config"
	"strconv"
	"strings"
)

func CheckRequest(path, format string) error {
	splitReq := strings.Split(path, "/")
	splitFormat := strings.Split(format, "/")
	if len(splitReq) != len(splitFormat) {
		return ErrHttpNotFound
	}
	return nil
}

func CheckMetricType(metricType, metricValue string) error {
	switch metricType {
	case MetricNameGauge:
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			return fmt.Errorf("%w: %s", ErrHttpBadRequest, MessageNotFloat)
		}
	case MetricNameCounter:
		if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
			return fmt.Errorf("%w: %s", ErrHttpBadRequest, MessageNotInteger)
		}
	default:
		return fmt.Errorf("%w: %s", ErrHttpBadRequest, MessageNotSupported)
	}
	return nil
}
