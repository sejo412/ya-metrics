package models

import "time"

// Metrics specific constants.
const (
	MetricKindGauge                string = "gauge"
	MetricKindCounter              string = "counter"
	MetricPathPostPrefix           string = "update"
	MetricPathPostsPrefix                 = MetricPathPostPrefix + "s"
	MetricPathGetPrefix            string = "value"
	MetricNamePollCount            string = "PollCount"
	MetricNameRandomValue          string = "RandomValue"
	MetricNameTotalMemory          string = "TotalMemory"
	MetricNameFreeMemory           string = "FreeMemory"
	MetricNamePrefixCPUUtilization string = "CPUutilization"
	PingPath                       string = "ping"
)

// HTTP headers.
const (
	HTTPHeaderContentTypeApplicationJSON     string = "application/json"
	HTTPHeaderContentTypeApplicationTextHTML string = "text/html"
	HTTPHeaderEncodingGzip                   string = "gzip"
	HTTPHeaderContentType                    string = "Content-Type"
	HTTPHeaderContentEncoding                string = "Content-Encoding"
	HTTPHeaderAcceptEncoding                 string = "Accept-Encoding"
	HTTPHeaderSign                           string = "HashSHA256"
)

// Ancillary constants.
const (
	base10        int = 10
	metricBitSize int = 64
)

// Retries constants.
const (
	RetryMaxRetries int           = 3
	RetryInitDelay  time.Duration = 1 * time.Second
	RetryDeltaDelay time.Duration = 2 * time.Second
)

const (
	TotalCountMetrics int = 43 // total metrics count
)
