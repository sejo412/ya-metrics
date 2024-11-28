package storage

import "time"

type Metric struct {
	Kind      string
	Name      string
	Value     any
	Timestamp time.Time
}

type Storage interface {
	Add(metric Metric)
	Last(Metric) (metric Metric)
	LastAll() (metrics []Metric)
}
