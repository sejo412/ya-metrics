package main

import (
	"fmt"
	"github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/server"
	"github.com/sejo412/ya-metrics/internal/storage"
	"html/template"
	"io"
	"log"
	"net/http"
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
	store.Add(metric)
}

func getValue(w http.ResponseWriter, r *http.Request) {
	if err := checkRequest(w, r, config.MetricPathGetFormat); err != nil {
		log.Print(err)
	}
	metric := parseGetValueRequest(r)
	store := r.Context().Value("store").(*storage.MemoryStorage)
	sum, err := server.GetMetricSum(store, metric)
	if err != nil {
		http.Error(w, config.MessageNotFound, http.StatusNotFound)
		return
	}
	if _, err = io.WriteString(w, fmt.Sprintf("%s=%v",
		metric.Name, sum)); err != nil {
		log.Print(err)
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	store := r.Context().Value("store").(*storage.MemoryStorage)
	LastMetrics := store.LastAll()
	tmpl, err := template.New("index").Parse(index)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
	err = tmpl.Execute(w, LastMetrics)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		log.Print(err)
		return
	}
}
