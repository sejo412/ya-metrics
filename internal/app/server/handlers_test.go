package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sejo412/ya-metrics/internal/config"
	logger2 "github.com/sejo412/ya-metrics/internal/logger"
	m "github.com/sejo412/ya-metrics/internal/models"
	"github.com/sejo412/ya-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var cfg = config.ServerConfig{
	Address:         ":8080",
	StoreInterval:   30,
	FileStoragePath: "/tmp/testing_metrics.json",
	Restore:         true,
}

const notFound = "404 page not found"

func Test_handleUpdate(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request string
		header  http.Header
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
			name:    "bad gauge",
			request: "/update/gauge/Frees/preved",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", m.ErrHTTPBadRequest, m.MessageNotFloat),
			},
		},
		{
			name:    "bad counter",
			request: "/update/counter/Frees/10.55",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", m.ErrHTTPBadRequest, m.MessageNotInteger),
			},
		},
		{
			name:    "bad type",
			request: "/update/preved/Frees/10",
			want: want{
				code:     http.StatusBadRequest,
				response: fmt.Sprintf("%s: %s", m.ErrHTTPBadRequest, m.MessageNotSupported),
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
			name:    "POST 405",
			request: "",
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "",
			},
		},
		{
			name:    "other not found",
			request: "/updateeee/qwe/asd",
			want: want{
				code:     http.StatusNotFound,
				response: notFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			defer func() {
				_ = logger.Sync()
			}()
			sugar := logger.Sugar()
			lm := logger2.NewMiddleware(sugar)
			store := storage.NewMemoryStorage()

			r := NewRouterWithConfig(&config.Options{
				Config:  cfg,
				Storage: store,
			}, lm)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodPost, tt.request, nil, nil)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}
func Test_getIndex(t *testing.T) {
	// notFound := "404 page not found"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			defer func() {
				_ = logger.Sync()
			}()
			sugar := logger.Sugar()
			lm := logger2.NewMiddleware(sugar)
			store := storage.NewMemoryStorage()

			r := NewRouterWithConfig(&config.Options{
				Config:  cfg,
				Storage: store,
			}, lm)
			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodGet, tt.request, nil, nil)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Contains(t, body, tt.want.response, tt.name)
		})
	}
}
func testRequest(t *testing.T, ts *httptest.Server, method, path string, header http.Header,
	body io.Reader) (*http.Response, string) {
	ctx := context.TODO()
	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	req.Header = header
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

func Test_postUpdateJSON(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name    string
		request string
		header  http.Header
		body    io.Reader
		want    want
	}{
		{
			name:    "not json",
			request: "/update/",
			header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			body: bytes.NewBuffer([]byte(`foo=bar`)),
			want: want{
				code:     http.StatusBadRequest,
				response: m.ErrHTTPBadRequest.Error(),
			},
		},
		{
			name:    "invalid json",
			request: "/update/",
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			body: bytes.NewBuffer([]byte(`{"value": "foo""}`)),
			want: want{
				code:     http.StatusBadRequest,
				response: m.ErrHTTPBadRequest.Error(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			defer func() {
				_ = logger.Sync()
			}()
			sugar := logger.Sugar()
			lm := logger2.NewMiddleware(sugar)
			store := storage.NewMemoryStorage()

			r := NewRouterWithConfig(&config.Options{
				Config:  cfg,
				Storage: store,
			}, lm)

			ts := httptest.NewServer(r)
			defer ts.Close()
			resp, body := testRequest(t, ts, http.MethodPost, tt.request, tt.header, tt.body)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
			assert.Equal(t, tt.want.response, body, tt.name)
		})
	}
}
