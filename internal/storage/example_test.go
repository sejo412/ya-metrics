package storage

import (
	"context"
	"fmt"

	"github.com/sejo412/ya-metrics/internal/models"
)

var exampleMemoryStorage = NewMemoryStorage()

func ExampleMemoryStorage_Upsert() {
	err := exampleMemoryStorage.Upsert(context.Background(), models.Metric{
		Kind:  models.MetricKindGauge,
		Name:  "metric1",
		Value: "99.99",
	})

	fmt.Println(err)

	// Output:
	// <nil>
}

func ExampleMemoryStorage_Get() {
	m, _ := exampleMemoryStorage.Get(context.Background(), models.MetricKindGauge, "metric1")
	fmt.Println(m.Value)

	// Output:
	// 99.99
}

func ExampleMemoryStorage_MassUpsert() {
	metrics := []models.Metric{
		models.Metric{
			Kind:  models.MetricKindGauge,
			Name:  "metric1",
			Value: "88.99",
		},
		models.Metric{
			Kind:  models.MetricKindCounter,
			Name:  "metric2",
			Value: "2",
		},
	}
	err := exampleMemoryStorage.MassUpsert(context.Background(), metrics)
	fmt.Println(err)

	// Output:
	// <nil>
}

func ExampleMemoryStorage_GetAll() {
	metrics, _ := exampleMemoryStorage.GetAll(context.Background())
	for _, m := range metrics {
		fmt.Printf("%s=%s\n", m.Name, m.Value)
	}

	// Output:
	// metric1=88.99
	// metric2=2
}
