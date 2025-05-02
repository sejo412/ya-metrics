package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"sync"

	"github.com/sejo412/ya-metrics/internal/models"
)

// MemoryStorage is backend for RAM.
type MemoryStorage struct {
	metrics map[string]models.Metric
	mutex   sync.Mutex
}

// NewMemoryStorage returns new MemoryStorage object.
func NewMemoryStorage() *MemoryStorage {
	metrics := make(map[string]models.Metric, models.TotalCountMetrics)
	return &MemoryStorage{metrics: metrics}
}

// Open not implemented for RAM.
func (s *MemoryStorage) Open(ctx context.Context, opts Options) error {
	return nil
}

func (s *MemoryStorage) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.metrics = nil
}

// Ping not implemented for RAM.
func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

// Init not implemented for RAM.
func (s *MemoryStorage) Init(ctx context.Context) error {
	return nil
}

// Upsert inserts or updates metric.
func (s *MemoryStorage) Upsert(ctx context.Context, metric models.Metric) error {
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

// MassUpsert inserts or updates slice of metrics.
func (s *MemoryStorage) MassUpsert(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		if err := s.Upsert(ctx, metric); err != nil {
			return err
		}
	}
	return nil
}

// Get returns metric by name.
// kind not implemented for RAM storage.
func (s *MemoryStorage) Get(ctx context.Context, kind, name string) (models.Metric, error) {
	// kind not used in this implementation, because name is "primary key" for MemoryStorage
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if metric, ok := s.metrics[name]; ok {
		return metric, nil
	}
	return models.Metric{}, models.ErrHTTPNotFound
}

// GetAll returns slice of all metrics.
func (s *MemoryStorage) GetAll(ctx context.Context) ([]models.Metric, error) {
	s.mutex.Lock()
	metrics := make([]models.Metric, 0, len(s.metrics))
	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}
	s.mutex.Unlock()
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics, nil
}

// Flush saves metrics to destination.
func (s *MemoryStorage) Flush(ctx context.Context, dst io.Writer) error {
	metrics, _ := s.GetAll(ctx)
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

// Load loads metrics from source.
func (s *MemoryStorage) Load(ctx context.Context, src io.Reader) error {
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
		if err = s.Upsert(ctx, *res); err != nil {
			return fmt.Errorf("error add or update metric %s: %w", res.Name, err)
		}
	}
	return nil
}
