package models

import (
	"runtime"
	"strconv"

	pb "github.com/sejo412/ya-metrics/proto"
)

// RuntimeMetricsMap returns mapping for runtime metrics.
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

// PSMetricsCPU returns map with CPU utilization.
func PSMetricsCPU(c []float64) map[string]float64 {
	result := make(map[string]float64)
	for i, value := range c {
		result[MetricNamePrefixCPUUtilization+strconv.Itoa(i)] = value
	}
	return result
}

// ConvertV1ToV2 converts V1 api to V2 for backward compatibility.
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

// ConvertV2ToV1 converts V2 api to V1 for backward compatibility.
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

// ConvertPbToV1 converts protobuf type to V1.
func ConvertPbToV1(m *pb.Metric) Metric {
	var mType string
	var value string

	switch *m.Type {
	case pb.MType_GAUGE:
		mType = "gauge"
		value = strconv.FormatFloat(m.GetValue(), 'f', -1, metricBitSize)
	case pb.MType_COUNTER:
		mType = "counter"
		value = strconv.FormatInt(m.GetDelta(), base10)
	}
	return Metric{
		Kind:  mType,
		Name:  *m.Id,
		Value: value,
	}
}

// ConvertPbsToV1s converts protobuf type to V1 slices.
func ConvertPbsToV1s(m []*pb.Metric) []Metric {
	result := make([]Metric, len(m))
	for i, metric := range m {
		result[i] = ConvertPbToV1(metric)
	}
	return result
}

// ConvertV1ToPb converts V1 metric to protobuf type.
func ConvertV1ToPb(m Metric) (*pb.Metric, error) {
	res := new(pb.Metric)
	res.Id = &m.Name
	res.Type, res.Delta, res.Value = nil, nil, nil
	switch m.Kind {
	case "counter":
		mType := pb.MType_COUNTER
		delta, err := strconv.Atoi(m.Value)
		value := int64(delta)
		if err != nil {
			return nil, ErrNotInteger
		}
		res.Type = &mType
		res.Delta = &value
	case "gauge":
		mType := pb.MType_GAUGE
		value, err := strconv.ParseFloat(m.Value, metricBitSize)
		if err != nil {
			return nil, ErrNotFloat
		}
		res.Type = &mType
		res.Value = &value
	default:
		return nil, ErrNotSupported
	}
	return res, nil
}

// ConvertV1sToPbs converts V1 metric slice to protobuf slice type.
func ConvertV1sToPbs(m []Metric) ([]*pb.Metric, error) {
	var err error
	result := make([]*pb.Metric, len(m))
	for i, metric := range m {
		result[i], err = ConvertV1ToPb(metric)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
