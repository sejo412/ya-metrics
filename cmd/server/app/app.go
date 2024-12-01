package app

import (
	"fmt"
	"github.com/sejo412/ya-metrics/internal/domain"
	"math"
	"strconv"
)

type Storage interface {
	AddOrUpdate(domain.Metric) error
	Get(string) (domain.Metric, error)
	GetAll() []domain.Metric
}

// CheckMetricKind returns error if metric value does not match kind
func CheckMetricKind(metric domain.Metric) error {
	_, err := getMetricValueString(metric)
	switch err {
	case domain.ErrNotFloat:
		return fmt.Errorf("%w: %s", domain.ErrHTTPBadRequest, domain.MessageNotFloat)
	case domain.ErrNotInteger:
		return fmt.Errorf("%w: %s", domain.ErrHTTPBadRequest, domain.MessageNotInteger)
	case domain.ErrNotSupported:
		return fmt.Errorf("%w: %s", domain.ErrHTTPBadRequest, domain.MessageNotSupported)
	}
	return nil
}

func UpdateMetric(st Storage, metric domain.Metric) error {
	return st.AddOrUpdate(metric)
}

func GetMetricValue(st Storage, name string) (string, error) {
	metric, err := st.Get(name)
	if err != nil {
		return "", err
	}
	return getMetricValueString(metric)
}

func GetAllMetricValues(st Storage) map[string]string {
	result := make(map[string]string)
	for _, metric := range st.GetAll() {
		value, err := getMetricValueString(metric)
		if err == nil {
			result[metric.Name] = value
		}
	}
	return result
}

// RoundFloatToString round float and convert it to string (trims trailing zeroes)
func roundFloatToString(val float64) string {
	ratio := math.Pow(10, float64(3))
	res := math.Round(val*ratio) / ratio
	return strconv.FormatFloat(res, 'f', -1, 64)
}

func getMetricValueString(metric domain.Metric) (string, error) {
	switch metric.Kind {
	case domain.MetricKindGauge:
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return "", domain.ErrNotFloat
		}
		return roundFloatToString(v), nil
	case domain.MetricKindCounter:
		v, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return "", domain.ErrNotInteger
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "", domain.ErrNotSupported
	}
}
