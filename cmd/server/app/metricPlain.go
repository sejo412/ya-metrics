package app

import "github.com/sejo412/ya-metrics/internal/models"

func UpdateMetric(st Storage, metric models.Metric) error {
	return st.AddOrUpdate(metric)
}
