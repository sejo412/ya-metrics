package server

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/sejo412/ya-metrics/internal/config"
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
	ctx := context.Background()
	if err := st.AddOrUpdate(ctx, m); err != nil {
		return nil, err
	}
	return GetMetricJSON(st, metric.MType, metric.ID)
}

// UpdateMetricsFromJSON updates metrics from incoming JSON slice
func UpdateMetricsFromJSON(st config.Storage, req []byte) error {
	parsedMetrics, err := ParsePostRequestJSONSlice(req)
	if err != nil {
		return err
	}
	res := make([]models.Metric, 0, len(parsedMetrics))
	for _, metric := range parsedMetrics {
		m, err := models.ConvertV2ToV1(&metric)
		if err != nil {
			return err
		}
		res = append(res, *m)
	}
	ctx := context.Background()
	if err := st.MassAddOrUpdate(ctx, res); err != nil {
		return err
	}
	return nil
}

// GetMetricJSON return JSON representation metric by name
func GetMetricJSON(st config.Storage, kind, name string) ([]byte, error) {
	ctx := context.Background()
	metric, err := st.Get(ctx, kind, name)
	if err != nil {
		return nil, err
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

// ParsePostRequestJSONSlice coverts incoming json to []MetricV2
func ParsePostRequestJSONSlice(request []byte) ([]models.MetricV2, error) {
	metrics := make([]models.MetricV2, 0)
	if err := json.Unmarshal(request, &metrics); err != nil {
		return metrics, models.ErrHTTPBadRequest
	}
	return metrics, nil
}
