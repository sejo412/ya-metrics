package models

import "time"

// shared constants
const (
	MetricKindGauge       string = "gauge"
	MetricKindCounter     string = "counter"
	MetricPathPostPrefix  string = "update"
	MetricPathPostsPrefix        = MetricPathPostPrefix + "s"
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
	HTTPHeaderSign                           string = "HashSHA256"
)

const (
	base10        int = 10
	metricBitSize int = 64
)

const (
	RetryMaxRetries int           = 3
	RetryInitDelay  time.Duration = 1 * time.Second
	RetryDeltaDelay time.Duration = 2 * time.Second
)
