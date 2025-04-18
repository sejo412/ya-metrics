package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/models"
)

// CheckMetricKind returns error if metric value does not match kind.
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

// GetMetricValue returns metric value by name.
func GetMetricValue(st config.Storage, name string) (string, error) {
	metric, err := st.Get(context.TODO(), "", name)
	if err != nil {
		return "", err
	}
	return models.GetMetricValueString(metric)
}

// GetAllMetricValues returns all metrics.
func GetAllMetricValues(st config.Storage) map[string]string {
	ctx := context.Background()
	result := make(map[string]string)
	metrics, err := st.GetAll(ctx)
	if err != nil {
		return nil
	}
	for _, metric := range metrics {
		value, err := models.GetMetricValueString(metric)
		if err == nil {
			result[metric.Name] = value
		}
	}
	return result
}

// FlushingMetrics saves metrics to file.
func FlushingMetrics(st config.Storage, file string, interval int) {
	for {
		f, err := os.Create(file)
		if err != nil {
			log.Printf("error create file %s: %v\n", file, err)
			return
		}
		if err = st.Flush(context.TODO(), f); err != nil {
			log.Printf("Error flushing metrics: %s", err.Error())
		}
		err = f.Close()
		if err != nil {
			log.Printf("error closing file %s: %v\n", file, err)
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
