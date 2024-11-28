package server

import (
	"fmt"
	. "github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/storage"
	"strconv"
)

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

func GetMetricSum(store *storage.MemoryStorage, metric storage.Metric) (any, error) {
	var sumGauge float64 = 0
	var sumCounter int64 = 0
	var found = false

	for _, m := range store.Metrics {
		if m.Kind == metric.Kind && m.Name == metric.Name {
			found = true
			switch m.Kind {
			case MetricNameGauge:
				mFloat, err := strconv.ParseFloat(fmt.Sprint(m.Value), 64)
				if err != nil {
					return nil, err
				}
				sumGauge += mFloat
			case MetricNameCounter:
				mInt, err := strconv.ParseInt(fmt.Sprint(m.Value), 10, 64)
				if err != nil {
					return nil, err
				}
				sumCounter += mInt
			}
		}
	}
	if !found {
		return nil, ErrHttpNotFound
	}
	if metric.Kind == MetricNameGauge {
		return sumGauge, nil
	} else if metric.Kind == MetricNameCounter {
		return sumCounter, nil
	}
	return nil, fmt.Errorf("%w", ErrHttpNotFound)
}
