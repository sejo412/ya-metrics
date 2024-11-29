package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/storage"
	"net/http"
	"strconv"
)

func checkRequest(w http.ResponseWriter, r *http.Request, format string) error {
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

func parsePostUpdateRequest(r *http.Request) storage.Metric {
	return storage.Metric{
		Kind:  chi.URLParam(r, "kind"),
		Name:  chi.URLParam(r, "name"),
		Value: chi.URLParam(r, "value"),
	}
}

func parseGetValueRequest(r *http.Request) storage.Metric {
	return storage.Metric{
		Kind: chi.URLParam(r, "kind"),
		Name: chi.URLParam(r, "name"),
	}
}

func checkMetricType(metricKind, metricValue string) error {
	switch metricKind {
	case config.MetricNameGauge:
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			return fmt.Errorf("%w: %s", config.ErrHTTPBadRequest, config.MessageNotFloat)
		}
	case config.MetricNameCounter:
		if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
			return fmt.Errorf("%w: %s", config.ErrHTTPBadRequest, config.MessageNotInteger)
		}
	default:
		return fmt.Errorf("%w: %s", config.ErrHTTPBadRequest, config.MessageNotSupported)
	}
	return nil
}
