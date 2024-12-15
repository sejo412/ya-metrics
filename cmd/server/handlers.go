package main

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sejo412/ya-metrics/cmd/server/app"
	"github.com/sejo412/ya-metrics/internal/models"
)

var index = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Index</title>
</head>
<body>
	{{- range $k, $v := . }} 
	{{ $k }}={{ $v }}
	<br>
	{{- end }}
</body>
</html>
`

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func postUpdate(w http.ResponseWriter, r *http.Request) {
	metric := models.Metric{
		Kind:  chi.URLParam(r, "kind"),
		Name:  chi.URLParam(r, "name"),
		Value: chi.URLParam(r, "value"),
	}
	store := r.Context().Value("store").(app.Storage)
	if err := app.UpdateMetric(store, metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%v: add or update metric %s", err, metric.Name)
	}
}

func getValue(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	store := r.Context().Value("store").(app.Storage)
	value, err := app.GetMetricValue(store, name)
	switch {
	case errors.Is(err, models.ErrHTTPNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, models.ErrHTTPInternalServerError):
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%v: get value %s", err, name)
		return
	}
	if _, err = io.WriteString(w, value); err != nil {
		log.Printf("%v: write to response writer", err)
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value("store").(app.Storage)
	metrics := app.GetAllMetricValues(store)
	tmpl, err := template.New("index").Parse(index)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
	err = tmpl.Execute(w, metrics)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
}

func WithLogging(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 200,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		sugar.Infow(
			"incoming request",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size)
	}

	return http.HandlerFunc(fn)
}

func postUpdateJSON(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(models.HTTPHeaderContentType) != models.HTTPHeaderContentTypeApplicationJSON {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()
	store := r.Context().Value("store").(app.Storage)
	resp, err := app.UpdateMetricFromJSON(store, buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getMetricJSON(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(models.HTTPHeaderContentType) != models.HTTPHeaderContentTypeApplicationJSON {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()
	store := r.Context().Value("store").(app.Storage)
	metric, err := app.ParsePostRequestJSON(buf.Bytes())
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
	}
	resp, err := app.GetMetricJSON(store, metric.MType, metric.ID)
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
	}
	w.Header().Set(models.HTTPHeaderContentType, "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
