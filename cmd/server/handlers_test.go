package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	. "github.com/sejo412/ya-metrics/internal/config"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_handleUpdate(t *testing.T) {
	notFound := "404 page not found"
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "ok gauge",
			request: "/update/gauge/testGauge/10.909",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name:    "ok counter",
			request: "/update/counter/testCounter/10",
			want: want{
				code:     http.StatusOK,
				response: "",
			},
		},
		{
			name: "bad gauge",
			//		pattern: "/update/{kind}/Frees/{value}",
			request: "/update/gauge/Frees/preved",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", ErrHTTPBadRequest, MessageNotFloat),
			},
		},
		{
			name:    "bad counter",
			request: "/update/counter/Frees/10.55",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", ErrHTTPBadRequest, MessageNotInteger),
			},
		},
		{
			name:    "bad type",
			request: "/update/preved/Frees/10",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", ErrHTTPBadRequest, MessageNotSupported),
			},
		},
		{
			name:    "too short",
			request: "/update/gauge/10",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
		{
			name:    "too long",
			request: "/update/gauge/Frees/subfree/10",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
		{
			name:    "index not found",
			request: "",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
		{
			name:    "other not found",
			request: "/updateeee/qwe/asd",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := "/update/{kind}/{name}/{value}"
			r := chi.NewRouter()
			store := storage.NewMemoryStorage()
			r.Use(middleware.WithValue("store", store))
			r.Handle(http.MethodPost+" "+pattern, http.HandlerFunc(postUpdate))
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodPost, tt.request, nil)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}
func Test_getIndex(t *testing.T) {
	//notFound := "404 page not found"
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "ok index",
			request: "/",
			want: want{
				code:     http.StatusOK,
				response: "<!DOCTYPE html>",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := "/value/{kind}/{name}"
			pattern = "/"
			r := chi.NewRouter()
			store := storage.NewMemoryStorage()
			r.Use(middleware.WithValue("store", store))
			r.Handle(http.MethodGet+" "+pattern, http.HandlerFunc(getIndex))
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodGet, tt.request, nil)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			//assert.Equal(t, tt.want.response, body, tt.name)
			assert.Contains(t, body, tt.want.response, tt.name)
		})
	}
}
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()
	result, _ := strings.CutSuffix(string(respBody), "\n")
	return resp, result
}
