package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/sejo412/ya-metrics/internal/models"
)

type MemoryStorage struct {
	mutex   sync.Mutex
	metrics map[string]models.Metric
}

func NewMemoryStorage() *MemoryStorage {
	metrics := make(map[string]models.Metric)
	return &MemoryStorage{metrics: metrics}
}

func (s *MemoryStorage) Open(ctx context.Context, opts Options) error {
	return nil
}

func (s *MemoryStorage) Close() {
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) Init(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) AddOrUpdate(ctx context.Context, metric models.Metric) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if metric.Kind == models.MetricKindCounter {
		if m, ok := s.metrics[metric.Name]; ok {
			currentInt, err := strconv.Atoi(m.Value)
			if err != nil {
				return fmt.Errorf("could not convert saved metric '%s' to int", metric.Name)
			}
			newInt, err := strconv.Atoi(metric.Value)
			if err != nil {
				return fmt.Errorf("could not convert new metric '%s' to int", metric.Name)
			}
			currentInt += newInt
			metric.Value = strconv.Itoa(currentInt)
		}
	}
	s.metrics[metric.Name] = metric
	return nil
}

func (s *MemoryStorage) MassAddOrUpdate(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := s.AddOrUpdate(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, kind, name string) (models.Metric, error) {
	// kind not used in this implementation, because name is "primary key" for MemoryStorage
	if metric, ok := s.metrics[name]; ok {
		return metric, nil
	}
	return models.Metric{}, models.ErrHTTPNotFound
}

func (s *MemoryStorage) GetAll(ctx context.Context) ([]models.Metric, error) {
	metrics := make([]models.Metric, 0, len(s.metrics))
	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (s *MemoryStorage) Flush(dst io.Writer) error {
	metrics, _ := s.GetAll(context.TODO())
	for _, metric := range metrics {
		m, err := models.ConvertV1ToV2(&metric)
		if err != nil {
			return err
		}
		err = json.NewEncoder(dst).Encode(m)
		if err != nil {
			return fmt.Errorf("error encode metric %s: %w", metric, err)
		}
	}
	return nil
}

func (s *MemoryStorage) Load(src io.Reader) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		m := models.MetricV2{}
		err := json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			return fmt.Errorf("error unmarshal metric %s: %w", scanner.Text(), err)
		}
		res, err := models.ConvertV2ToV1(&m)
		if err != nil {
			return err
		}
		if err = s.AddOrUpdate(context.TODO(), *res); err != nil {
			return fmt.Errorf("error add or update metric %s: %w", res.Name, err)
		}
	}
	return nil
}
