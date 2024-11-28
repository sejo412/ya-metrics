package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sejo412/ya-metrics/internal/server"
	"github.com/sejo412/ya-metrics/internal/storage"
	"net/http"
	"path/filepath"
	"time"
)

func checkRequest(w http.ResponseWriter, r *http.Request, format string) error {
	// skip check metric kind for method GET
	if r.Method == http.MethodGet {
		return nil
	}
	path := filepath.Clean(r.URL.Path)
	metricKind := chi.URLParam(r, "kind")
	metricValue := chi.URLParam(r, "value")

	if err := server.CheckMetricType(metricKind, metricValue); err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return fmt.Errorf("%w: %s %s", err, http.MethodPost, path)
	}
	return nil
}

func parsePostUpdateRequest(r *http.Request) storage.Metric {
	return storage.Metric{
		Kind:      chi.URLParam(r, "kind"),
		Name:      chi.URLParam(r, "name"),
		Value:     chi.URLParam(r, "value"),
		Timestamp: time.Now(),
	}
}

func parseGetValueRequest(r *http.Request) storage.Metric {
	return storage.Metric{
		Kind: chi.URLParam(r, "kind"),
		Name: chi.URLParam(r, "name"),
	}
}
