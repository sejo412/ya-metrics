package app

import (
	"encoding/json"
	"strconv"

	"github.com/sejo412/ya-metrics/cmd/server/config"
	"github.com/sejo412/ya-metrics/internal/models"
)

// UpdateMetricFromJSON updates metric from incoming json
func UpdateMetricFromJSON(st config.Storage, req []byte) ([]byte, error) {
	var (
		metric models.MetricV2
		err    error
	)
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
func GetMetricJSON(st config.Storage, kind, name string) ([]byte, error) {
	metric, err := st.Get("", name)
	if err != nil {
		return nil, err
	}
	// kind is dummy for MemoryStorage. in feature implementations (with other storages) this workaround will be removed
	if metric.Kind != kind {
		return nil, models.ErrHTTPNotFound
	}
	m, err := models.ConvertV1ToV2(&metric)
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
