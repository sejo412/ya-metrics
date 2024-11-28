package storage

type MemoryStorage struct {
	Metrics []Metric
}

func NewMemoryStorage() *MemoryStorage {
	metrics := make([]Metric, 0)
	return &MemoryStorage{metrics}
}

func (s *MemoryStorage) Add(metric Metric) {
	s.Metrics = append(s.Metrics, metric)
}

func (s *MemoryStorage) Last(metric Metric) Metric {
	for i := len(s.Metrics) - 1; i >= 0; i-- {
		if metric.Kind == s.Metrics[i].Kind && metric.Name == s.Metrics[i].Name {
			return s.Metrics[i]
		}
	}
	return Metric{}
}

func (s *MemoryStorage) LastAll() []Metric {
	result := make([]Metric, 0)
	tmp := make(map[string]any)
	for i := len(s.Metrics) - 1; i >= 0; i-- {
		if _, ok := tmp[s.Metrics[i].Name]; !ok {
			tmp[s.Metrics[i].Name] = s.Metrics[i].Value
		}
	}
	for k, v := range tmp {
		result = append(result, Metric{Name: k, Value: v})
	}
	return result
}
