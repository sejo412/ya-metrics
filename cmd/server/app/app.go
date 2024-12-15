package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/sejo412/ya-metrics/internal/models"
)

const base10 float64 = 10

type Storage interface {
	AddOrUpdate(models.Metric) error
	Get(kind string, name string) (models.Metric, error)
	GetAll() []models.Metric
}

// CheckMetricKind returns error if metric value does not match kind
func CheckMetricKind(metric models.Metric) error {
	_, err := getMetricValueString(metric)
	switch {
	case errors.Is(err, models.ErrNotFloat):
		return fmt.Errorf("%w: %s", models.ErrHTTPBadRequest, models.MessageNotFloat)
	case errors.Is(err, models.ErrNotInteger):
		return fmt.Errorf("%w: %s", models.ErrHTTPBadRequest, models.MessageNotInteger)
	case errors.Is(err, models.ErrNotSupported):
		return fmt.Errorf("%w: %s", models.ErrHTTPBadRequest, models.MessageNotSupported)
	}
	return nil
}

func UpdateMetric(st Storage, metric models.Metric) error {
	return st.AddOrUpdate(metric)
}

func GetMetricValue(st Storage, name string) (string, error) {
	metric, err := st.Get("", name)
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

/*
New api for post JSON
*/

// UpdateMetricFromJSON updates metric from incoming json
func UpdateMetricFromJSON(st Storage, req []byte) ([]byte, error) {
	var metric models.MetricV2
	var err error
	metric, err = ParsePostRequestJSON(req)
	if err != nil {
		return nil, err
	}
	m := models.Metric{
		Kind: metric.MType,
		Name: metric.ID,
	}
	switch metric.MType {
	case models.MetricKindGauge:
		m.Value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	case models.MetricKindCounter:
		m.Value = strconv.FormatInt(*metric.Delta, 10)
	default:
		return nil, models.ErrNotSupported
	}

	if err = CheckMetricKind(m); err != nil {
		return nil, err
	}
	if err = st.AddOrUpdate(m); err != nil {
		return nil, err
	}
	return GetMetricJSON(st, metric.MType, metric.ID)
}

// GetMetricJSON return JSON representation metric by name
func GetMetricJSON(st Storage, kind, name string) ([]byte, error) {
	metric, err := st.Get("", name)
	if err != nil {
		return nil, err
	}
	// kind is dummy for MemoryStorage. in feature implementations (with other storages) this workaround will be removed
	if metric.Kind != kind {
		return nil, models.ErrHTTPNotFound
	}
	m := models.MetricV2{
		ID:    name,
		MType: kind,
	}
	switch metric.Kind {
	case models.MetricKindGauge:
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return nil, models.ErrNotFloat
		}
		m.Value = &v
	case models.MetricKindCounter:
		v, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return nil, models.ErrNotInteger
		}
		m.Delta = &v
	}
	return json.Marshal(metric)
}

// ParsePostRequestJSON converts incoming json to MetricV2 type
func ParsePostRequestJSON(request []byte) (models.MetricV2, error) {
	metrics := models.MetricV2{}
	if err := json.Unmarshal(request, &metrics); err != nil {
		return metrics, models.ErrHTTPBadRequest
	}
	return metrics, nil
}
