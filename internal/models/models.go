package models

import (
	"math"
	"runtime"
	"strconv"
)

const base10 int = 10
const metricBitSize int = 64

type Metric struct {
	Kind  string
	Name  string
	Value string
}

type MetricV2 struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func RuntimeMetricsMap(r *runtime.MemStats) map[string]interface{} {
	return map[string]interface{}{
		"Alloc":         r.Alloc,
		"TotalAlloc":    r.TotalAlloc,
		"Sys":           r.Sys,
		"Lookups":       r.Lookups,
		"Mallocs":       r.Mallocs,
		"Frees":         r.Frees,
		"HeapAlloc":     r.HeapAlloc,
		"HeapSys":       r.HeapSys,
		"HeapIdle":      r.HeapIdle,
		"HeapInuse":     r.HeapInuse,
		"HeapReleased":  r.HeapReleased,
		"HeapObjects":   r.HeapObjects,
		"StackInuse":    r.StackInuse,
		"StackSys":      r.StackSys,
		"MSpanInuse":    r.MSpanInuse,
		"MSpanSys":      r.MSpanSys,
		"MCacheInuse":   r.MCacheInuse,
		"MCacheSys":     r.MCacheSys,
		"BuckHashSys":   r.BuckHashSys,
		"GCSys":         r.GCSys,
		"OtherSys":      r.OtherSys,
		"NextGC":        r.NextGC,
		"LastGC":        r.LastGC,
		"PauseTotalNs":  r.PauseTotalNs,
		"NumGC":         r.NumGC,
		"NumForcedGC":   r.NumForcedGC,
		"GCCPUFraction": r.GCCPUFraction,
	}
}

func ConvertMetricToV2(m *Metric) (*MetricV2, error) {
	res := &MetricV2{
		ID:    m.Name,
		MType: m.Kind,
	}
	switch m.Kind {
	case "counter":
		v, err := strconv.ParseInt(m.Value, base10, metricBitSize)
		if err != nil {
			return nil, ErrNotInteger
		}
		res.Delta = &v
	case "gauge":
		v, err := strconv.ParseFloat(m.Value, metricBitSize)
		if err != nil {
			return nil, ErrNotFloat
		}
		res.Value = &v
	}
	return res, nil
}

func ConvertV2ToMetric(m *MetricV2) (*Metric, error) {
	metric := &Metric{
		Kind: m.MType,
		Name: m.ID,
	}
	switch m.MType {
	case "counter":
		metric.Value = strconv.FormatInt(*m.Delta, base10)
	case "gauge":
		metric.Value = strconv.FormatFloat(*m.Value, 'f', -1, metricBitSize)
	}
	return metric, nil
}

// RoundFloatToString round float and convert it to string (trims trailing zeroes)
func RoundFloatToString(val float64) string {
	ratio := math.Pow(float64(base10), float64(3))
	res := math.Round(val*ratio) / ratio
	return strconv.FormatFloat(res, 'f', -1, 64)
}

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
