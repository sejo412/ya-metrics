package main

import (
	"github.com/sejo412/ya-metrics/cmd/server/app"
	"github.com/sejo412/ya-metrics/internal/domain"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

var index = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Index</title>
</head>
<body>
	{{- range $k, $v := . }} 
	{{ $v.Name }}={{ $v.Value }}
	<br>
	{{- end }}
</body>
</html>
`

type Storage interface {
	AddOrUpdate(domain.Metric) error
	Get(string) (domain.Metric, error)
	GetAll() []domain.Metric
}

func postUpdate(w http.ResponseWriter, r *http.Request) {
	if err := app.CheckRequest(w, r); err != nil {
		log.Print(err)
	}
	if _, err := io.WriteString(w, ""); err != nil {
		log.Print(err)
	}
	metric := app.ParsePostUpdateRequest(r)
	store := r.Context().Value("store").(Storage)
	if err := addOrUpdate(store, metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%v: add or update metric %s", err, metric.Name)
	}
}

func getValue(w http.ResponseWriter, r *http.Request) {
	if err := app.CheckRequest(w, r); err != nil {
		log.Print(err)
	}
	req := app.ParseGetValueRequest(r)
	//store := r.Context().Value("store").(*storage.MemoryStorage)
	store := r.Context().Value("store").(Storage)
	metric, err := get(store, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var value string
	switch metric.Kind {
	case domain.MetricKindCounter:
		v, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("%v: parse value %s", err, metric.Value)
			return
		}
		value = strconv.FormatInt(v, 10)
	case domain.MetricKindGauge:
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("%v: parse value %s", err, metric.Value)
			return
		}
		value = app.RoundFloatToString(v)
	}

	if _, err := io.WriteString(w, value); err != nil {
		log.Print(err)
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value("store").(Storage)
	metrics := getAll(store)
	for i, metric := range metrics {
		if metric.Kind == domain.MetricKindGauge {
			vFloat, err := strconv.ParseFloat(metric.Value, 64)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			metrics[i].Value = app.RoundFloatToString(vFloat)
		}
	}
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

func getAll(st Storage) []domain.Metric {
	return st.GetAll()
}

func addOrUpdate(st Storage, metric domain.Metric) error {
	if err := st.AddOrUpdate(metric); err != nil {
		return err
	}
	return nil
}

func get(st Storage, name string) (domain.Metric, error) {
	metric, err := st.Get(name)
	if err != nil {
		return domain.Metric{}, err
	}
	return metric, nil
}
