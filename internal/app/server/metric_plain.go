package server

import (
	"context"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/models"
)

// UpdateMetric inserts or updates MetricV1
func UpdateMetric(st config.Storage, metric models.Metric) error {
	return st.Upsert(context.TODO(), metric)
}
