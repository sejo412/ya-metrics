package models

import (
	"math"
	"strconv"
)

// Metric describes metric object.
type Metric struct {
	// Kind - gauge or counter
	Kind string
	// Name - metric name
	Name string
	// Value - metric value
	Value string
}

// GetMetricValueString returns string of metric value.
func GetMetricValueString(metric Metric) (string, error) {
	switch metric.Kind {
	case MetricKindGauge:
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return "", ErrNotFloat
		}
		return RoundFloatToString(v), nil
	case MetricKindCounter:
		v, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return "", ErrNotInteger
		}
		return strconv.FormatInt(v, 10), nil
	default:
		return "", ErrNotSupported
	}
}

// RoundFloatToString round float and convert it to string.
// Trims trailing zeroes.
func RoundFloatToString(val float64) string {
	ratio := math.Pow(float64(base10), float64(3))
	res := math.Round(val*ratio) / ratio
	return strconv.FormatFloat(res, 'f', -1, 64)
}
