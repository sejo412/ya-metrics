package main

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sejo412/ya-metrics/internal/utils"

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

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// POST and gzip
		if r.Method == http.MethodPost &&
			r.Header.Get(models.HTTPHeaderContentEncoding) == models.HTTPHeaderEncodingGzip {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer func() {
				_ = r.Body.Close()
			}()
			data, err := utils.Decompress(buf.Bytes())
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(data))
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
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
	cfg := r.Context().Value("config").(Config)
	if cfg.StoreInterval == 0 {
		if err := store.Flush(cfg.FileStoragePath); err != nil {
			log.Printf("%v: flush store %s", err, cfg.FileStoragePath)
		}
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

	w.Header().Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationTextHTML)
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, metrics)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
	resp := buf.Bytes()
	if r.Header.Get(models.HTTPHeaderAcceptEncoding) == models.HTTPHeaderEncodingGzip {
		resp, err = utils.Compress(resp)
		if err == nil {
			w.Header().Set(models.HTTPHeaderContentEncoding, models.HTTPHeaderEncodingGzip)
		}
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
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

	data := buf.Bytes()

	store := r.Context().Value("store").(app.Storage)
	resp, err := app.UpdateMetricFromJSON(store, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := r.Context().Value("config").(Config)
	if cfg.StoreInterval == 0 {
		if err := store.Flush(cfg.FileStoragePath); err != nil {
			log.Printf("%v: flush store %s", err, cfg.FileStoragePath)
		}
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
		return
	}
	resp, err := app.GetMetricJSON(store, metric.MType, metric.ID)
	if err != nil {
		http.Error(w, models.ErrHTTPNotFound.Error(), http.StatusNotFound)
		return
	}
	if r.Header.Get(models.HTTPHeaderAcceptEncoding) == models.HTTPHeaderEncodingGzip {
		resp, err = utils.Compress(resp)
		if err == nil {
			w.Header().Set(models.HTTPHeaderContentEncoding, models.HTTPHeaderEncodingGzip)
		}
	}
	w.Header().Set(models.HTTPHeaderContentType, "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
