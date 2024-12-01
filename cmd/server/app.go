package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sejo412/ya-metrics/internal/domain"
	"math"
	"net/http"
	"strconv"
)

// checkRequest wrapper for checkMetricType function for POST method
func checkRequest(w http.ResponseWriter, r *http.Request) error {
	// skip check metric kind for method GET
	if r.Method == http.MethodGet {
		return nil
	}
	metricKind := chi.URLParam(r, "kind")
	metricValue := chi.URLParam(r, "value")

	if err := checkMetricType(metricKind, metricValue); err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return fmt.Errorf("%w: %s %s", err, http.MethodPost, r.URL.Path)
	}
	return nil
}

// parsePostUpdateRequest parses POST params to Metric type
func parsePostUpdateRequest(r *http.Request) domain.Metric {
	return domain.Metric{
		Kind:  chi.URLParam(r, "kind"),
		Name:  chi.URLParam(r, "name"),
		Value: chi.URLParam(r, "value"),
	}
}

// parseGetValueRequest parses GET for request Metric
func parseGetValueRequest(r *http.Request) domain.Metric {
	return domain.Metric{
		Kind: chi.URLParam(r, "kind"),
		Name: chi.URLParam(r, "name"),
	}
}

// checkMetricType returns error if metric value does not match kind
func checkMetricType(metricKind, metricValue string) error {
	switch metricKind {
	case domain.MetricKindGauge:
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			return fmt.Errorf("%w: %s", domain.ErrHTTPBadRequest, domain.MessageNotFloat)
		}
	case domain.MetricKindCounter:
		if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
			return fmt.Errorf("%w: %s", domain.ErrHTTPBadRequest, domain.MessageNotInteger)
		}
	default:
		return fmt.Errorf("%w: %s", domain.ErrHTTPBadRequest, domain.MessageNotSupported)
	}
	return nil
}

// roundFloatToString round float and convert it to string (trims trailing zeroes)
func roundFloatToString(val float64) string {
	ratio := math.Pow(10, float64(3))
	res := math.Round(val*ratio) / ratio
	return strconv.FormatFloat(res, 'f', -1, 64)
}
