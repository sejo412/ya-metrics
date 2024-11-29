package main

import (
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/storage"
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

func postUpdate(w http.ResponseWriter, r *http.Request) {
	if err := checkRequest(w, r, config.MetricPathPostFormat); err != nil {
		log.Print(err)
	}
	if _, err := io.WriteString(w, ""); err != nil {
		log.Print(err)
	}
	metric := parsePostUpdateRequest(r)
	store := r.Context().Value("store").(*storage.MemoryStorage)
	if err := store.AddOrUpdate(metric); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%v: add or update metric %s", err, metric.Name)
	}
}

func getValue(w http.ResponseWriter, r *http.Request) {
	if err := checkRequest(w, r, config.MetricPathGetFormat); err != nil {
		log.Print(err)
	}
	req := parseGetValueRequest(r)
	store := r.Context().Value("store").(*storage.MemoryStorage)
	metric, err := store.Get(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var value string
	switch metric.Kind {
	case config.MetricNameCounter:
		v, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("%v: parse value %s", err, metric.Value)
			return
		}
		value = strconv.FormatInt(v, 10)
	case config.MetricNameGauge:
		v, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Printf("%v: parse value %s", err, metric.Value)
			return
		}
		value = roundFloatToString(v)
	}

	if _, err := io.WriteString(w, value); err != nil {
		log.Print(err)
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value("store").(*storage.MemoryStorage)
	Metrics := store.GetAll()
	for i, metric := range Metrics {
		if metric.Kind == config.MetricNameGauge {
			vFloat, err := strconv.ParseFloat(metric.Value, 64)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			Metrics[i].Value = roundFloatToString(vFloat)
		}
	}
	tmpl, err := template.New("index").Parse(index)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
	err = tmpl.Execute(w, Metrics)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
}
