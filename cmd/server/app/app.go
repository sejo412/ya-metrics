package app

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/sejo412/ya-metrics/internal/models"
)

type Storage interface {
	AddOrUpdate(models.Metric) error
	Get(kind string, name string) (models.Metric, error)
	GetAll() []models.Metric
	Flush(dst io.Writer) error
	Load(src io.Reader) error
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

func FlushingMetrics(st Storage, file string, interval int) {
	for {
		f, err := os.Create(file)
		if err != nil {
			log.Printf("error create file %s: %v\n", file, err)
			return
		}
		if err = st.Flush(f); err != nil {
			log.Printf("Error flushing metrics: %s", err.Error())
		}
		err = f.Close()
		if err != nil {
			log.Printf("error closing file %s: %v\n", file, err)
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
