package config

const (
	MetricTypeGauge   string = "float64"
	MetricTypeCounter string = "int64"
	MetricNameGauge   string = "gauge"
	MetricNameCounter string = "counter"
)

const (
	MetricPathPostPrefix string = "/upload"
	MetricPathPostFormat        = MetricPathPostPrefix + "/%s/%s/%s"
)
