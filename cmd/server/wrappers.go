package main

import (
	"fmt"
	. "github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/server"
	"net/http"
	"path/filepath"
	"strings"
)

func checkPostRequest(w http.ResponseWriter, r *http.Request) error {
	path := filepath.Clean(r.URL.Path)
	if err := server.CheckRequest(path, MetricPathPostFormat); err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusNotFound)
		return fmt.Errorf("%w: %s", err, path)
	}

	splitReq := strings.Split(path, "/")
	if err := server.CheckMetricType(splitReq[2], splitReq[4]); err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusBadRequest)
		return fmt.Errorf("%w: %s %s", err, http.MethodPost, path)
	}
	return nil
}
