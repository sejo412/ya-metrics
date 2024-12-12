package app

import (
	"fmt"
	"math"
	"strconv"

	"github.com/sejo412/ya-metrics/internal/models"
)

const base10 float64 = 10

type Storage interface {
	AddOrUpdate(models.Metric) error
	Get(string) (models.Metric, error)
	GetAll() []models.Metric
}

// CheckMetricKind returns error if metric value does not match kind
func CheckMetricKind(metric models.Metric) error {
	_, err := getMetricValueString(metric)
	switch err {
	case models.ErrNotFloat:
		return fmt.Errorf("%w: %s", models.ErrHTTPBadRequest, models.MessageNotFloat)
	case models.ErrNotInteger:
		return fmt.Errorf("%w: %s", models.ErrHTTPBadRequest, models.MessageNotInteger)
	case models.ErrNotSupported:
		return fmt.Errorf("%w: %s", models.ErrHTTPBadRequest, models.MessageNotSupported)
	}
	return nil
}

func UpdateMetric(st Storage, metric models.Metric) error {
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
	ratio := math.Pow(base10, float64(3))
	res := math.Round(val*ratio) / ratio
	return strconv.FormatFloat(res, 'f', -1, 64)
}

func getMetricValueString(metric models.Metric) (string, error) {
	switch metric.Kind {
	case models.MetricKindGauge:
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return "", models.ErrNotFloat
		}
		return roundFloatToString(v), nil
	case models.MetricKindCounter:
		v, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return "", models.ErrNotInteger
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "", models.ErrNotSupported
	}
}
