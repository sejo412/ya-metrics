package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	listenScheme  string = "http"
	listenAddress string = "0.0.0.0"
	listenPort    string = "8080"
)

type Metrics interface {
	Add()
	Replace()
}

type MemStorage struct {
	Metrics []Metric
}

type Metric struct {
	Name string
	Type any
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	checkRequest(w, r)

}

func (m Metric) Add() {

}

func (m Metric) Replace() {}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/", handleUpdate)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", listenAddress, listenPort), mux); err != nil {
		panic(err)
	}

}

func checkRequest(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)
	splittedReq := strings.Split(path, "/")
	if len(splittedReq) != 5 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	checkMetricType(w, splittedReq[2], splittedReq[4])
	return
}

func checkMetricType(w http.ResponseWriter, t, v string) {
	switch t {
	case "gauge":
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			http.Error(w, "Bad request: not float64", http.StatusBadRequest)
		} else {
			return
		}
	case "counter":
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			http.Error(w, "Bad request: not int64", http.StatusBadRequest)
		} else {
			return
		}
	default:
		http.Error(w, "Bad request: Unknown type", http.StatusBadRequest)
	}
}
