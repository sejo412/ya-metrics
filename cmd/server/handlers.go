package main

import (
	"log"
	"net/http"
)

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if err := checkPostRequest(w, r); err != nil {
		log.Print(err)
	}
	//log.Printf("%s %s", r.Method, r.URL.Path)
	w.Write()
}
