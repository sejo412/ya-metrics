package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/sejo412/ya-metrics/internal/models"
)

type Storage interface {
	AddOrUpdate(models.Metric) error
	Get(kind string, name string) (models.Metric, error)
	GetAll() []models.Metric
	Flush(file string) error
	Load(file string) error
}

// CheckMetricKind returns error if metric value does not match kind
func CheckMetricKind(metric models.Metric) error {
	_, err := models.GetMetricValueString(metric)
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
	return models.GetMetricValueString(metric)
}

func GetAllMetricValues(st Storage) map[string]string {
	result := make(map[string]string)
	for _, metric := range st.GetAll() {
		value, err := models.GetMetricValueString(metric)
		if err == nil {
			result[metric.Name] = value
		}
	}
	return result
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

	if err := CheckMetricKind(m); err != nil {
		return nil, err
	}
	if err := st.AddOrUpdate(m); err != nil {
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
	m, err := models.ConvertMetricToV2(&metric)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// ParsePostRequestJSON converts incoming json to MetricV2 type
func ParsePostRequestJSON(request []byte) (models.MetricV2, error) {
	metrics := models.MetricV2{}
	if err := json.Unmarshal(request, &metrics); err != nil {
		return metrics, models.ErrHTTPBadRequest
	}
	return metrics, nil
}

func FlushingMetrics(st Storage, file string, interval int) {
	for {
		if err := st.Flush(file); err != nil {
			log.Printf("Error flushing metrics: %s", err.Error())
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
