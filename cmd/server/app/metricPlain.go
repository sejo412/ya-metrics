package app

import (
	"github.com/sejo412/ya-metrics/cmd/server/config"
	"github.com/sejo412/ya-metrics/internal/models"
)

func UpdateMetric(st config.Storage, metric models.Metric) error {
	return st.AddOrUpdate(metric)
}
