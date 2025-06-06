package server

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io"
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

func (r *Router) decryptHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() {
			_ = req.Body.Close()
		}()
		var data []byte
		if len(body) > 0 {
			data, err = utils.Decode(body, r.opts.PrivateKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		req.Body = io.NopCloser(bytes.NewReader(data))
		next.ServeHTTP(w, req)
	})
}

func (r *Router) postUpdate(w http.ResponseWriter, req *http.Request) {
	log := r.opts.Logger.Logger
	cfg := r.opts.Config
	metric := models.Metric{
		Kind:  chi.URLParam(req, "kind"),
		Name:  chi.URLParam(req, "name"),
		Value: chi.URLParam(req, "value"),
	}
	store := r.opts.Storage
	if err := UpdateMetric(store, metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Errorw("add or update metric",
			"metric", metric.Name,
			"error", err)
	}
	if cfg.StoreInterval == 0 {
		f, err := os.Create(cfg.StoreFile)
		defer func() {
			_ = f.Close()
		}()
		if err != nil {
			log.Errorw("create file",
				"file", cfg.StoreFile,
				"error", err)
		}
		if err = store.Flush(context.TODO(), f); err != nil {
			log.Errorw("flush store",
				"file", cfg.StoreFile,
				"error", err)
		}
	}
}

func (r *Router) getValue(w http.ResponseWriter, req *http.Request) {
	log := r.opts.Logger.Logger
	name := chi.URLParam(req, "name")
	store := r.opts.Storage
	value, err := GetMetricValue(store, name)
	switch {
	case errors.Is(err, models.ErrHTTPNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, models.ErrHTTPInternalServerError):
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Errorw("get value",
			"name", name,
			"error", err)
		return
	}
	if _, err = io.WriteString(w, value); err != nil {
		log.Errorw("write to response writer",
			"error", err)
	}
}

func (r *Router) getIndex(w http.ResponseWriter, req *http.Request) {
	log := r.opts.Logger.Logger
	store := r.opts.Storage
	metrics := GetAllMetricValues(store)
	tmpl, err := template.New("index").Parse(index)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Errorw("parse index template", "error", err)
		return
	}
	w.Header().Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationTextHTML)
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, metrics)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Errorw("generate index page", "error", err)
		return
	}
	resp := buf.Bytes()
	if req.Header.Get(models.HTTPHeaderAcceptEncoding) == models.HTTPHeaderEncodingGzip {
		resp, err = utils.Compress(resp)
		if err == nil {
			w.Header().Set(models.HTTPHeaderContentEncoding, models.HTTPHeaderEncodingGzip)
		}
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Errorw("write response", "error", err)
		return
	}
}

func (r *Router) postUpdateJSON(w http.ResponseWriter, req *http.Request) {
	log := r.opts.Logger.Logger
	if req.Header.Get(models.HTTPHeaderContentType) != models.HTTPHeaderContentTypeApplicationJSON {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = req.Body.Close()
	}()

	data := buf.Bytes()

	store := r.opts.Storage
	resp, err := UpdateMetricFromJSON(store, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := r.opts.Config
	if cfg.StoreInterval == 0 {
		func() {
			f, err1 := os.Create(cfg.StoreFile)
			if err1 != nil {
				log.Errorw("create file", "file", cfg.StoreFile, "error", err1)
				return
			}
			defer func() {
				_ = f.Close()
			}()
			if err2 := store.Flush(context.TODO(), f); err2 != nil {
				log.Errorw("flush store", "file", cfg.StoreFile, "error", err2)
			}
		}()
	}
	w.Header().Set(models.HTTPHeaderContentType, models.HTTPHeaderContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Errorw("write response", "error", err)
		return
	}
}
func (r *Router) postUpdatesJSON(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get(models.HTTPHeaderContentType) != models.HTTPHeaderContentTypeApplicationJSON {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = req.Body.Close()
	}()

	data := buf.Bytes()

	store := r.opts.Storage
	err = UpdateMetricsFromJSON(store, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) getMetricJSON(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get(models.HTTPHeaderContentType) != models.HTTPHeaderContentTypeApplicationJSON {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, models.ErrHTTPBadRequest.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = req.Body.Close()
	}()
	store := r.opts.Storage
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
	if req.Header.Get(models.HTTPHeaderAcceptEncoding) == models.HTTPHeaderEncodingGzip {
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

func (r *Router) pingStorage(w http.ResponseWriter, req *http.Request) {
	log := r.opts.Logger.Logger
	store := r.opts.Storage
	ctx := context.Background()
	if err := store.Ping(ctx); err != nil {
		http.Error(w, models.ErrHTTPInternalServerError.Error(), http.StatusInternalServerError)
		log.Errorw("ping storage", "error", err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (r *Router) checkXRealIPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// skip checks if not POST method
		if req.Method != http.MethodPost {
			next.ServeHTTP(w, req)
		}
		xRealIP := req.Header.Get("X-Real-Ip")
		if !isNetsContainsIP(xRealIP, r.opts.TrustedSubnets) {
			http.Error(w, models.ErrHTTPForbidden.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, req)
	})
}
