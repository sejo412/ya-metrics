package storage

type Metric struct {
	Kind  string
	Name  string
	Value string
}

type Storage interface {
	AddOrUpdate(Metric) error
	Get(string) (Metric, error)
	GetAll() []Metric
}
