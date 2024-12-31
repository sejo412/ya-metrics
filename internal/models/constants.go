package models

// shared constants
const (
	MetricKindGauge       string = "gauge"
	MetricKindCounter     string = "counter"
	MetricPathPostPrefix  string = "update"
	MetricPathGetPrefix   string = "value"
	MetricNamePollCount   string = "PollCount"
	MetricNameRandomValue string = "RandomValue"
	PingPath              string = "ping"
)

const (
	HTTPHeaderContentTypeApplicationJSON     string = "application/json"
	HTTPHeaderContentTypeApplicationTextHTML string = "text/html"
	HTTPHeaderEncodingGzip                   string = "gzip"
	HTTPHeaderContentType                    string = "Content-Type"
	HTTPHeaderContentEncoding                string = "Content-Encoding"
	HTTPHeaderAcceptEncoding                 string = "Accept-Encoding"
)

const (
	base10        int = 10
	metricBitSize int = 64
)
