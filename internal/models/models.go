package models

import (
	"runtime"
	"strconv"
)

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

func PSMetricsCPU(c []float64) map[string]float64 {
	result := make(map[string]float64)
	for i, value := range c {
		result[MetricNamePrefixCPUUtilization+strconv.Itoa(i)] = value
	}
	return result
}

func ConvertV1ToV2(m *Metric) (*MetricV2, error) {
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

func ConvertV2ToV1(m *MetricV2) (*Metric, error) {
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
