package models

// MetricV2 describes metric object.
type MetricV2 struct {
	// ID - metrics id.
	ID string `json:"id"`
	// MType - gauge or counter.
	MType string `json:"type"`
	// Delta.
	Delta *int64 `json:"delta,omitempty"`
	// Value - metrics value.
	Value *float64 `json:"value,omitempty"`
}
