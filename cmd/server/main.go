package main

import (
	"fmt"
	"net/http"
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
