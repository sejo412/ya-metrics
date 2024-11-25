package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	messageNotFloat     string = "Bad request: not a float"
	messageNotInt       string = "Bad request: not a int"
	messageNotSupported string = "Bad request: not a supported type"
	messageNotFound     string = "Not found"
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
	if err := checkRequest(w, r); err != nil {
		log.Print(err)
	}
	log.Printf("%s %s", r.Method, r.URL.Path)
}

func (m Metric) Add() {

}

func (m Metric) Replace() {}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/", handleUpdate)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, messageNotFound, http.StatusNotFound)
	})

	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", listenAddress, listenPort), mux); err != nil {
		panic(err)
	}
}

func checkRequest(w http.ResponseWriter, r *http.Request) error {
	path := filepath.Clean(r.URL.Path)
	splittedReq := strings.Split(path, "/")
	if len(splittedReq) != 5 {
		http.Error(w, messageNotFound, http.StatusNotFound)
		return fmt.Errorf("%s", messageNotFound)
	}
	if err := checkMetricType(w, splittedReq[2], splittedReq[4]); err != nil {
		return err
	}
	return nil
}

func checkMetricType(w http.ResponseWriter, t, v string) error {
	switch t {
	case "gauge":
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			http.Error(w, messageNotFloat, http.StatusBadRequest)
			return fmt.Errorf("%s", messageNotFloat)
		}
	case "counter":
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			http.Error(w, messageNotInt, http.StatusBadRequest)
			return fmt.Errorf("%s", messageNotInt)
		}
	default:
		http.Error(w, messageNotSupported, http.StatusBadRequest)
		return fmt.Errorf("%s", messageNotSupported)
	}
	return nil
}
