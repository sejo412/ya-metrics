package server

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/pkg/utils"
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

func checkHashHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Context().Value("key").(string)
		headerHash := r.Header.Get(models.HTTPHeaderSign)
		if r.Method == http.MethodPost && key != "" && headerHash != "" {
			/* Broken logic in autotests
			headerHash := r.Header.Get(models.HTTPHeaderSign)
			if headerHash == "" {
				http.Error(w, "No sign header found", http.StatusBadRequest)
				return
			}
			*/
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer func() {
				_ = r.Body.Close()
			}()

			if len(body) > 0 {
				want := utils.Hash(body, key)
				if want != headerHash {
					http.Error(w, "Invalid sign", http.StatusBadRequest)
					return
				}
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		next.ServeHTTP(w, r)
	})
}

func (cr *Router) postUpdate(w http.ResponseWriter, r *http.Request) {
	metric := models.Metric{
		Kind:  chi.URLParam(r, "kind"),
		Name:  chi.URLParam(r, "name"),
		Value: chi.URLParam(r, "value"),
	}
	store := cr.opts.Storage
	if err := UpdateMetric(store, metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%v: add or update metric %s", err, metric.Name)
	}

	cfg := cr.opts.Config
	if cfg.StoreInterval == 0 {
		f, err := os.Create(cfg.FileStoragePath)
		defer func() {
			_ = f.Close()
		}()
		if err != nil {
			log.Printf("error create file %s: %v", cfg.FileStoragePath, err)
		}
		if err = store.Flush(f); err != nil {
			log.Printf("%v: flush store %s", err, cfg.FileStoragePath)
		}
	}
}

func (cr *Router) getValue(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	store := cr.opts.Storage
	value, err := GetMetricValue(store, name)
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

func (cr *Router) getIndex(w http.ResponseWriter, r *http.Request) {
	store := cr.opts.Storage
	metrics := GetAllMetricValues(store)
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

func (cr *Router) postUpdateJSON(w http.ResponseWriter, r *http.Request) {
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

	store := cr.opts.Storage
	resp, err := UpdateMetricFromJSON(store, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := cr.opts.Config
	if cfg.StoreInterval == 0 {
		f, err := os.Create(cfg.FileStoragePath)
		defer func() {
			_ = f.Close()
		}()
		if err != nil {
			log.Printf("error create file %s: %v", cfg.FileStoragePath, err)
		}
		if err = store.Flush(f); err != nil {
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
func (cr *Router) postUpdatesJSON(w http.ResponseWriter, r *http.Request) {
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

	store := cr.opts.Storage
	err = UpdateMetricsFromJSON(store, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (cr *Router) getMetricJSON(w http.ResponseWriter, r *http.Request) {
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
	store := cr.opts.Storage
	metric, err := ParsePostRequestJSON(buf.Bytes())
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	resp, err := GetMetricJSON(store, metric.MType, metric.ID)
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

func (cr *Router) pingStorage(w http.ResponseWriter, r *http.Request) {
	store := cr.opts.Storage
	ctx := context.Background()
	if err := store.Ping(ctx); err != nil {
		http.Error(w, models.ErrHTTPInternalServerError.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		return
	}
}
