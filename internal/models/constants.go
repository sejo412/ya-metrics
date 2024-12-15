package models

// shared constants
const (
	MetricKindGauge       string = "gauge"
	MetricKindCounter     string = "counter"
	MetricPathPostPrefix  string = "update"
	MetricPathGetPrefix   string = "value"
	MetricNamePollCount   string = "PollCount"
	MetricNameRandomValue string = "RandomValue"
)

const (
	HTTPHeaderContentTypeApplicationJSON string = "application/json"
	HTTPHeaderContentType                string = "Content-Type"
)
