package config

const (
	MetricTypeGauge   string = "float64"
	MetricTypeCounter string = "int64"
	MetricNameGauge   string = "gauge"
	MetricNameCounter string = "counter"
)

const (
	MetricPathPostPrefix string = "/upload"
	MetricPathGetPrefix  string = "/value"
	MetricPathPostFormat        = MetricPathPostPrefix + "/%s/%s/%s"
	MetricPathGetFormat         = MetricPathGetPrefix + "/%s/%s"
)
