package models

// MetricV2 describes metric object.
type MetricV2 struct {
	// Delta.
	Delta *int64 `json:"delta,omitempty"`
	// Value - metrics value.
	Value *float64 `json:"value,omitempty"`
	// ID - metrics id.
	ID string `json:"id"`
	// MType - gauge or counter.
	MType string `json:"type"`
}
